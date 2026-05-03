package repositories

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/infra/dao"
)

type TreasuryAPIDAO interface {
	FetchRates(ctx context.Context, currency string, startDate, endDate string) (*dao.TreasuryRateResponse, error)
}

type TreasuryRateRepository struct {
	dao TreasuryAPIDAO
}

func NewTreasuryRateRepository(dao TreasuryAPIDAO) *TreasuryRateRepository {
	return &TreasuryRateRepository{dao: dao}
}

func (p *TreasuryRateRepository) GetRate(ctx context.Context, targetCurrency string, transactionDate time.Time) (domain.CurrencyConversionRate, error) {
	startDate := transactionDate.AddDate(0, -6, 0).Format("2006-01-02")
	endDate := transactionDate.Format("2006-01-02")

	resp, err := p.dao.FetchRates(ctx, targetCurrency, startDate, endDate)
	if err != nil {
		return domain.CurrencyConversionRate{}, err
	}

	if len(resp.Data) == 0 {
		return domain.CurrencyConversionRate{}, domain.ErrNoConversionRate
	}

	record := resp.Data[0]
	rate, err := decimal.NewFromString(record.ExchangeRate)
	if err != nil {
		return domain.CurrencyConversionRate{}, err
	}

	rateDate, _ := time.Parse("2006-01-02", record.RecordDate)

	return domain.CurrencyConversionRate{
		TargetCurrency: targetCurrency,
		RateDate:       rateDate,
		ExchangeRate:   rate,
	}, nil
}
