package providers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/infra/dao"
	"wex/conversion_service/src/infra/repositories"
)

type MockTreasuryAPIDAO struct {
	mock.Mock
}

func (m *MockTreasuryAPIDAO) FetchRates(ctx context.Context, currency string, startDate, endDate string) (*dao.TreasuryRateResponse, error) {
	args := m.Called(ctx, currency, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dao.TreasuryRateResponse), args.Error(1)
}

type MockValkeyDAO struct {
	mock.Mock
}

func (m *MockValkeyDAO) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockValkeyDAO) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockValkeyDAO) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, value, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockValkeyDAO) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockValkeyDAO) SetStatus(ctx context.Context, id string, status string, expiration time.Duration) error {
	args := m.Called(ctx, id, status, expiration)
	return args.Error(0)
}

func (m *MockValkeyDAO) GetStatus(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func TestRateCacheProvider_GetRate_ValkeyHit(t *testing.T) {
	mockTreasuryDAO := new(MockTreasuryAPIDAO)
	treasuryRepo := repositories.NewTreasuryRateRepository(mockTreasuryDAO)

	db, _, _ := sqlmock.New()
	defer db.Close()
	postgresDAO := dao.NewPostgresDAO(db)
	postgresRepo := repositories.NewRatePostgresRepository(postgresDAO)

	mockValkeyDAO := new(MockValkeyDAO)
	valkeyRepo := repositories.NewValkeyRepository(mockValkeyDAO)

	provider := NewRateCacheProvider(treasuryRepo, postgresRepo, valkeyRepo)

	ctx := context.Background()
	date := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)
	targetCurrency := "BRL"
	valkeyKey := "rate:BRL:2023-10-27"

	cr := cachedRate{
		Rate: domain.CurrencyConversionRate{
			TargetCurrency: "BRL",
			RateDate:       date,
			ExchangeRate:   decimal.NewFromFloat(5.0),
		},
		ExpiresAt: time.Now().Add(1 * time.Hour), // Valid
	}
	crBytes, _ := json.Marshal(cr)

	mockValkeyDAO.On("Get", ctx, valkeyKey).Return(string(crBytes), nil)

	rate, err := provider.GetRate(ctx, targetCurrency, date)
	assert.NoError(t, err)
	assert.Equal(t, "5", rate.ExchangeRate.String())
	mockValkeyDAO.AssertExpectations(t)
}

func TestRateCacheProvider_GetRate_L2Hit(t *testing.T) {
	mockTreasuryDAO := new(MockTreasuryAPIDAO)
	treasuryRepo := repositories.NewTreasuryRateRepository(mockTreasuryDAO)

	db, sqlMock, _ := sqlmock.New()
	defer db.Close()
	postgresDAO := dao.NewPostgresDAO(db)
	postgresRepo := repositories.NewRatePostgresRepository(postgresDAO)

	mockValkeyDAO := new(MockValkeyDAO)
	valkeyRepo := repositories.NewValkeyRepository(mockValkeyDAO)

	provider := NewRateCacheProvider(treasuryRepo, postgresRepo, valkeyRepo)

	ctx := context.Background()
	date := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)
	targetCurrency := "BRL"
	valkeyKey := "rate:BRL:2023-10-27"

	mockValkeyDAO.On("Get", ctx, valkeyKey).Return("", errors.New("not found"))
	
	// Postgres will return hit
	rows := sqlmock.NewRows([]string{"target_currency", "rate_date", "exchange_rate", "created_at", "updated_at"}).
		AddRow("BRL", date, decimal.NewFromFloat(5.5), time.Now(), time.Now())
	sqlMock.ExpectQuery("SELECT target_currency").WithArgs(targetCurrency, "2023-10-27").WillReturnRows(rows)

	mockValkeyDAO.On("Set", ctx, valkeyKey, mock.Anything, 24*time.Hour).Return(nil)

	rate, err := provider.GetRate(ctx, targetCurrency, date)
	assert.NoError(t, err)
	assert.Equal(t, "5.5", rate.ExchangeRate.String())
	mockValkeyDAO.AssertExpectations(t)
}
