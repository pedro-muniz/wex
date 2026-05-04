package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wex/api_service/src/core/domain"
	"wex/api_service/src/core/ports"
)

// MockPayloadStore is a mock implementation of ports.PayloadStore
type MockPayloadStore struct {
	mock.Mock
}

func (m *MockPayloadStore) StorePayload(ctx context.Context, jobID uuid.UUID, payload domain.PurchaseTransaction) error {
	args := m.Called(ctx, jobID, payload)
	return args.Error(0)
}

func (m *MockPayloadStore) GetPayload(ctx context.Context, jobID uuid.UUID) (domain.PurchaseTransaction, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(domain.PurchaseTransaction), args.Error(1)
}

func (m *MockPayloadStore) UpdateStatus(ctx context.Context, jobID uuid.UUID, status domain.TransactionStatus) error {
	args := m.Called(ctx, jobID, status)
	return args.Error(0)
}

func (m *MockPayloadStore) GetStatus(ctx context.Context, jobID uuid.UUID) (domain.TransactionStatus, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(domain.TransactionStatus), args.Error(1)
}

func (m *MockPayloadStore) DeletePayload(ctx context.Context, jobID uuid.UUID) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockPayloadStore) SetRaw(ctx context.Context, key string, data string) error {
	args := m.Called(ctx, key, data)
	return args.Error(0)
}

func (m *MockPayloadStore) GetRaw(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// MockMessagePublisher is a mock implementation of ports.MessagePublisher
type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) PublishJob(ctx context.Context, jobID uuid.UUID) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishConversionRequest(ctx context.Context, jobID uuid.UUID, currency string) error {
	args := m.Called(ctx, jobID, currency)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishSyncRequest(ctx context.Context, jobID uuid.UUID) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func TestConversionProducerService(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	currency := "USD"

	t.Run("RequestConversion - success", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewConversionProducerService(mockStore, mockPub)

		mockPub.On("PublishConversionRequest", ctx, id, currency).Return(nil)

		err := service.RequestConversion(ctx, id, currency)

		assert.NoError(t, err)
		mockPub.AssertExpectations(t)
	})

	t.Run("RequestConversion - validation failure", func(t *testing.T) {
		mockPub := new(MockMessagePublisher)
		service := NewConversionProducerService(nil, mockPub)

		err := service.RequestConversion(ctx, id, "") // Empty currency fails validation

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidation))
		mockPub.AssertNotCalled(t, "PublishConversionRequest", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("RequestConversion - publish failure", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewConversionProducerService(mockStore, mockPub)

		expectedErr := errors.New("publish error")
		mockPub.On("PublishConversionRequest", ctx, id, currency).Return(expectedErr)

		err := service.RequestConversion(ctx, id, currency)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockPub.AssertExpectations(t)
	})

	t.Run("GetConversionResult - success", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		service := NewConversionProducerService(mockStore, nil)

		key := "key"
		expectedResult := "result"
		mockStore.On("GetRaw", ctx, key).Return(expectedResult, nil)

		result, err := service.GetConversionResult(ctx, key)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockStore.AssertExpectations(t)
	})

	t.Run("GetConversionResult - failure", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		service := NewConversionProducerService(mockStore, nil)

		key := "key"
		expectedErr := errors.New("store error")
		mockStore.On("GetRaw", ctx, key).Return("", expectedErr)

		_, err := service.GetConversionResult(ctx, key)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
	})
}

func TestTransactionProducerService(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	dto := ports.TransactionRequestDTO{
		Description:     "Test",
		TransactionDate: "2023-01-01",
		PurchaseAmount:  decimal.NewFromFloat(100.0),
	}

	t.Run("CreateTransaction - success", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		mockStore.On("StorePayload", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("domain.PurchaseTransaction")).Return(nil)
		mockPub.On("PublishJob", ctx, mock.AnythingOfType("uuid.UUID")).Return(nil)

		createdID, err := service.CreateTransaction(ctx, dto)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, createdID)
		mockStore.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("CreateTransaction - date parsing failure", func(t *testing.T) {
		service := NewTransactionProducerService(nil, nil)
		badDTO := dto
		badDTO.TransactionDate = "invalid"

		_, err := service.CreateTransaction(ctx, badDTO)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidation))
	})

	t.Run("CreateTransaction - validation failure", func(t *testing.T) {
		service := NewTransactionProducerService(nil, nil)
		badDTO := dto
		badDTO.PurchaseAmount = decimal.NewFromFloat(-1.0)

		_, err := service.CreateTransaction(ctx, badDTO)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrValidation))
	})

	t.Run("CreateTransaction - store failure", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		expectedErr := errors.New("store error")
		mockStore.On("StorePayload", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("domain.PurchaseTransaction")).Return(expectedErr)

		_, err := service.CreateTransaction(ctx, dto)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("CreateTransaction - publish failure", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		mockStore.On("StorePayload", ctx, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("domain.PurchaseTransaction")).Return(nil)
		expectedErr := errors.New("publish error")
		mockPub.On("PublishJob", ctx, mock.AnythingOfType("uuid.UUID")).Return(expectedErr)

		_, err := service.CreateTransaction(ctx, dto)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("GetTransactionStatus - success", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		mockPub.On("PublishSyncRequest", ctx, id).Return(nil)
		mockStore.On("GetStatus", ctx, id).Return(domain.StatusCompleted, nil)

		status, err := service.GetTransactionStatus(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, domain.StatusCompleted, status)
		mockStore.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("GetTransactionStatus - failure", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		mockPub.On("PublishSyncRequest", ctx, id).Return(nil)
		expectedErr := errors.New("get status error")
		mockStore.On("GetStatus", ctx, id).Return(domain.TransactionStatus(""), expectedErr)

		_, err := service.GetTransactionStatus(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("GetTransaction - success", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		tx := domain.PurchaseTransaction{ID: id}
		mockPub.On("PublishSyncRequest", ctx, id).Return(nil)
		mockStore.On("GetPayload", ctx, id).Return(tx, nil)

		result, err := service.GetTransaction(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, tx, result)
		mockStore.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("GetTransaction - failure", func(t *testing.T) {
		mockStore := new(MockPayloadStore)
		mockPub := new(MockMessagePublisher)
		service := NewTransactionProducerService(mockStore, mockPub)

		mockPub.On("PublishSyncRequest", ctx, id).Return(nil)
		expectedErr := errors.New("get payload error")
		mockStore.On("GetPayload", ctx, id).Return(domain.PurchaseTransaction{}, expectedErr)

		_, err := service.GetTransaction(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})
}
