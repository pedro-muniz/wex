package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wex/conversion_service/src/core/domain"
)

// Mock definitions
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.PurchaseTransaction), args.Error(1)
}

type MockRateProvider struct {
	mock.Mock
}

func (m *MockRateProvider) GetRate(ctx context.Context, currency string, date time.Time) (domain.CurrencyConversionRate, error) {
	args := m.Called(ctx, currency, date)
	return args.Get(0).(domain.CurrencyConversionRate), args.Error(1)
}

type MockPayloadStore struct {
	mock.Mock
}

func (m *MockPayloadStore) SetRaw(ctx context.Context, key string, data string) error {
	args := m.Called(ctx, key, data)
	return args.Error(0)
}

func (m *MockPayloadStore) GetRaw(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func TestGetConvertedTransaction(t *testing.T) {
	id := uuid.New()
	targetCurrency := "BRL"
	statusKey := fmt.Sprintf("conversion_status:%s:%s", id, targetCurrency)

	t.Run("Success", func(t *testing.T) {
		repo := new(MockRepo)
		rateProvider := new(MockRateProvider)
		payloadStore := new(MockPayloadStore)
		service := NewTransactionQueryService(repo, rateProvider, payloadStore)

		tx := domain.PurchaseTransaction{
			ID:     id,
			Amount: decimal.NewFromInt(100),
		}
		rate := domain.CurrencyConversionRate{
			ExchangeRate: decimal.NewFromFloat(5.0),
		}

		repo.On("GetByID", mock.Anything, id).Return(tx, nil)
		rateProvider.On("GetRate", mock.Anything, targetCurrency, mock.Anything).Return(rate, nil)
		payloadStore.On("SetRaw", mock.Anything, statusKey, string(domain.StatusCompleted)).Return(nil)

		resp, err := service.GetConvertedTransaction(context.Background(), id, targetCurrency)

		assert.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500).String(), resp.ConvertedAmount.String())
		assert.Contains(t, resp.Message, "Successfully converted")
		repo.AssertExpectations(t)
		rateProvider.AssertExpectations(t)
		payloadStore.AssertExpectations(t)
	})

	t.Run("Repo Failure", func(t *testing.T) {
		repo := new(MockRepo)
		rateProvider := new(MockRateProvider)
		payloadStore := new(MockPayloadStore)
		service := NewTransactionQueryService(repo, rateProvider, payloadStore)

		repo.On("GetByID", mock.Anything, id).Return(domain.PurchaseTransaction{}, errors.New("db error"))
		payloadStore.On("SetRaw", mock.Anything, statusKey, string(domain.StatusFailed)).Return(nil)

		_, err := service.GetConvertedTransaction(context.Background(), id, targetCurrency)

		assert.Error(t, err)
		payloadStore.AssertExpectations(t)
	})

	t.Run("RateProvider Failure", func(t *testing.T) {
		repo := new(MockRepo)
		rateProvider := new(MockRateProvider)
		payloadStore := new(MockPayloadStore)
		service := NewTransactionQueryService(repo, rateProvider, payloadStore)

		tx := domain.PurchaseTransaction{
			ID:     id,
			Amount: decimal.NewFromInt(100),
		}

		repo.On("GetByID", mock.Anything, id).Return(tx, nil)
		rateProvider.On("GetRate", mock.Anything, targetCurrency, mock.Anything).Return(domain.CurrencyConversionRate{}, errors.New("api error"))
		payloadStore.On("SetRaw", mock.Anything, statusKey, string(domain.StatusFailed)).Return(nil)

		_, err := service.GetConvertedTransaction(context.Background(), id, targetCurrency)

		assert.Error(t, err)
		payloadStore.AssertExpectations(t)
	})
}
