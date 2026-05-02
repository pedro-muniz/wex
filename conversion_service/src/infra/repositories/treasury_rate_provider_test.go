package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wex/conversion_service/src/infra/dao"
)

type MockTreasuryAPIDAO struct {
	mock.Mock
}

func (m *MockTreasuryAPIDAO) FetchRates(ctx context.Context, currency string, startDate, endDate string) (*dao.TreasuryRateResponse, error) {
	args := m.Called(ctx, currency, startDate, endDate)
	return args.Get(0).(*dao.TreasuryRateResponse), args.Error(1)
}

func TestTreasuryRateProvider(t *testing.T) {
	mockDAO := new(MockTreasuryAPIDAO)
	provider := NewTreasuryRateProvider(mockDAO)
	ctx := context.Background()
	date := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)

	t.Run("GetRate Success", func(t *testing.T) {
		mockResp := &dao.TreasuryRateResponse{
			Data: []struct {
				RecordDate      string `json:"record_date"`
				CountryCurrency string `json:"country_currency_desc"`
				ExchangeRate    string `json:"exchange_rate"`
				EffectiveDate   string `json:"effective_date"`
			}{
				{
					RecordDate:   "2023-09-30",
					ExchangeRate: "5.123",
				},
			},
		}

		mockDAO.On("FetchRates", ctx, "BRL", "2023-04-27", "2023-10-27").Return(mockResp, nil)

		rate, err := provider.GetRate(ctx, "BRL", date)
		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromFloat(5.123), rate.ExchangeRate)
		assert.Equal(t, "BRL", rate.TargetCurrency)
	})

	t.Run("GetRate No Rates Found", func(t *testing.T) {
		mockResp := &dao.TreasuryRateResponse{Data: []struct {
			RecordDate      string `json:"record_date"`
			CountryCurrency string `json:"country_currency_desc"`
			ExchangeRate    string `json:"exchange_rate"`
			EffectiveDate   string `json:"effective_date"`
		}{}}
		mockDAO.On("FetchRates", ctx, "XYZ", mock.Anything, mock.Anything).Return(mockResp, nil)

		_, err := provider.GetRate(ctx, "XYZ", date)
		assert.Error(t, err)
	})
}
