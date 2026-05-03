package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/infra/repositories"
)

type cachedRate struct {
	Rate      domain.CurrencyConversionRate `json:"rate"`
	ExpiresAt time.Time                     `json:"expires_at"` // Stale-while-revalidate meta
}

type RateCacheProvider struct {
	fallbackProvider *repositories.TreasuryRateRepository
	dbRepo           *repositories.RatePostgresRepository
	valkeyRepo       *repositories.ValkeyRepository
}

func NewRateCacheProvider(
	fallbackProvider *repositories.TreasuryRateRepository,
	dbRepo *repositories.RatePostgresRepository,
	valkeyRepo *repositories.ValkeyRepository,
) *RateCacheProvider {
	return &RateCacheProvider{
		fallbackProvider: fallbackProvider,
		dbRepo:           dbRepo,
		valkeyRepo:       valkeyRepo,
	}
}

func (s *RateCacheProvider) GetRate(ctx context.Context, targetCurrency string,
	transactionDate time.Time) (domain.CurrencyConversionRate, error) {
	dateStr := transactionDate.Format("2006-01-02")
	valkeyKey := fmt.Sprintf("rate:%s:%s", targetCurrency, dateStr)

	// 1. Check Valkey (L1) with exact date match
	if rate, hit := s.checkValkey(ctx, targetCurrency, transactionDate, valkeyKey); hit {
		return rate, nil
	}

	// 2. Check Postgres (L2) with 6 months rule
	if rate, hit := s.checkPostgres(ctx, targetCurrency, transactionDate, valkeyKey); hit {
		return rate, nil
	}

	// 3. Fallback to API with Distributed Lock with 6 months rule
	return s.fetchFromAPIWithLock(ctx, targetCurrency, transactionDate, valkeyKey, dateStr)
}

func (s *RateCacheProvider) checkValkey(ctx context.Context, targetCurrency string, transactionDate time.Time, valkeyKey string) (domain.CurrencyConversionRate, bool) {
	valStr, err := s.valkeyRepo.GetRaw(ctx, valkeyKey)
	if err == nil && valStr != "" {
		var cr cachedRate
		if err := json.Unmarshal([]byte(valStr), &cr); err == nil {
			if time.Now().Before(cr.ExpiresAt) {
				return cr.Rate, true // Hit and valid
			}
			// Hit but stale -> spawn background refresh, return stale data
			go s.fetchAndCacheAsync(context.Background(), targetCurrency,
				transactionDate, valkeyKey)
			return cr.Rate, true
		}
	}
	return domain.CurrencyConversionRate{}, false
}

func (s *RateCacheProvider) checkPostgres(ctx context.Context, targetCurrency string, transactionDate time.Time, valkeyKey string) (domain.CurrencyConversionRate, bool) {
	dbRate, err := s.dbRepo.FindByCurrencyAndDate(ctx, targetCurrency, transactionDate)
	if err == nil {
		s.cacheToValkey(ctx, valkeyKey, dbRate)
		return dbRate, true // Hit in L2
	}
	return domain.CurrencyConversionRate{}, false
}

func (s *RateCacheProvider) fetchFromAPIWithLock(ctx context.Context, targetCurrency string, transactionDate time.Time, valkeyKey, dateStr string) (domain.CurrencyConversionRate, error) {
	lockKey := fmt.Sprintf("lock:rate:%s:%s", targetCurrency, dateStr)
	acquired, err := s.valkeyRepo.SetNXRaw(ctx, lockKey, "locked", 10*time.Second)
	if err != nil {
		return domain.CurrencyConversionRate{}, err
	}

	if acquired {
		// We have the lock, fetch from API
		defer s.valkeyRepo.DelRaw(context.Background(), lockKey)

		apiRate, err := s.fallbackProvider.GetRate(ctx, targetCurrency, transactionDate)
		if err != nil {
			return domain.CurrencyConversionRate{}, err
		}

		// Persist to Postgres (Idempotent)
		_ = s.dbRepo.Upsert(ctx, apiRate)
		// Cache in Valkey
		s.cacheToValkey(ctx, valkeyKey, apiRate)

		return apiRate, nil
	}

	// Lock failed: wait and retry reading from Valkey once
	time.Sleep(500 * time.Millisecond)
	valStr, err := s.valkeyRepo.GetRaw(ctx, valkeyKey)
	if err == nil && valStr != "" {
		var cr cachedRate
		if err := json.Unmarshal([]byte(valStr), &cr); err == nil {
			return cr.Rate, nil
		}
	}

	return domain.CurrencyConversionRate{}, domain.ErrNoConversionRate
}

func (s *RateCacheProvider) cacheToValkey(ctx context.Context, key string, rate domain.CurrencyConversionRate) {
	// TTL logic: cache for 24h absolute, but stale after 12h
	cr := cachedRate{
		Rate:      rate,
		ExpiresAt: time.Now().Add(12 * time.Hour),
	}
	data, _ := json.Marshal(cr)
	_ = s.valkeyRepo.SetRawWithTTL(ctx, key, string(data), 24*time.Hour)
}

func (s *RateCacheProvider) fetchAndCacheAsync(ctx context.Context,
	targetCurrency string, transactionDate time.Time, valkeyKey string) {

	lockKey := fmt.Sprintf("lock:rate:%s:%s", targetCurrency, transactionDate.Format("2006-01-02"))
	acquired, _ := s.valkeyRepo.SetNXRaw(ctx, lockKey, "locked", 10*time.Second)
	if !acquired {
		return // Someone else is refreshing
	}
	defer s.valkeyRepo.DelRaw(ctx, lockKey)

	apiRate, err := s.fallbackProvider.GetRate(ctx, targetCurrency, transactionDate)
	if err != nil {
		log.Printf("Failed to refresh rate async: %v", err)
		return
	}

	_ = s.dbRepo.Upsert(ctx, apiRate)
	s.cacheToValkey(ctx, valkeyKey, apiRate)
}
