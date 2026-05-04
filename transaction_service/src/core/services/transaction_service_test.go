package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wex/transaction_service/src/core/domain"
)

// MockTransactionRepository is a mock implementation of ports.TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Save(ctx context.Context, tx domain.PurchaseTransaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.PurchaseTransaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(ctx context.Context, tx domain.PurchaseTransaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

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

func TestProcessTransaction(t *testing.T) {
	id := uuid.New()
	ctx := context.Background()
	tx := domain.PurchaseTransaction{
		ID:          id,
		Description: "Test",
		Amount:      decimal.NewFromFloat(100.0),
	}

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetPayload", ctx, id).Return(tx, nil)
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(nil)
		mockRepo.On("Save", ctx, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(nil)
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusCompleted
		})).Return(nil)

		err := service.ProcessTransaction(ctx, id)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failure - GetPayload", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		expectedErr := errors.New("get payload error")
		mockStore.On("GetPayload", ctx, id).Return(domain.PurchaseTransaction{}, expectedErr)
		mockStore.On("UpdateStatus", ctx, id, domain.StatusFailed).Return(nil)

		err := service.ProcessTransaction(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("failure - StorePayload PROCESSING", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetPayload", ctx, id).Return(tx, nil)
		expectedErr := errors.New("store error")
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(expectedErr)

		err := service.ProcessTransaction(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("failure - Repo Save", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetPayload", ctx, id).Return(tx, nil)
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(nil)
		
		expectedErr := errors.New("repo save error")
		mockRepo.On("Save", ctx, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(expectedErr)
		
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusFailed
		})).Return(nil)

		err := service.ProcessTransaction(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failure - StorePayload COMPLETED", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetPayload", ctx, id).Return(tx, nil)
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(nil)
		mockRepo.On("Save", ctx, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusProcessing
		})).Return(nil)
		
		expectedErr := errors.New("final store error")
		mockStore.On("StorePayload", ctx, id, mock.MatchedBy(func(p domain.PurchaseTransaction) bool {
			return p.Status == domain.StatusCompleted
		})).Return(expectedErr)

		err := service.ProcessTransaction(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetTransaction(t *testing.T) {
	id := uuid.New()
	ctx := context.Background()
	tx := domain.PurchaseTransaction{ID: id}

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		service := NewTransactionPersistenceService(mockRepo, nil)

		mockRepo.On("GetByID", ctx, id).Return(tx, nil)

		result, err := service.GetTransaction(ctx, id)

		assert.NoError(t, err)
		assert.Equal(t, tx, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		service := NewTransactionPersistenceService(mockRepo, nil)

		expectedErr := errors.New("not found")
		mockRepo.On("GetByID", ctx, id).Return(domain.PurchaseTransaction{}, expectedErr)

		_, err := service.GetTransaction(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSyncCache(t *testing.T) {
	id := uuid.New()
	ctx := context.Background()
	tx := domain.PurchaseTransaction{ID: id}

	t.Run("cache hit", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetStatus", ctx, id).Return(domain.StatusCompleted, nil)

		err := service.SyncCache(ctx, id)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	})

	t.Run("cache miss - success", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetStatus", ctx, id).Return(domain.TransactionStatus(""), errors.New("miss"))
		mockRepo.On("GetByID", ctx, id).Return(tx, nil)
		mockStore.On("StorePayload", ctx, id, tx).Return(nil)

		err := service.SyncCache(ctx, id)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cache miss - repo error", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetStatus", ctx, id).Return(domain.TransactionStatus(""), errors.New("miss"))
		expectedErr := errors.New("db error")
		mockRepo.On("GetByID", ctx, id).Return(domain.PurchaseTransaction{}, expectedErr)

		err := service.SyncCache(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("cache miss - store error", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockStore := new(MockPayloadStore)
		service := NewTransactionPersistenceService(mockRepo, mockStore)

		mockStore.On("GetStatus", ctx, id).Return(domain.TransactionStatus(""), errors.New("miss"))
		mockRepo.On("GetByID", ctx, id).Return(tx, nil)
		expectedErr := errors.New("redis error")
		mockStore.On("StorePayload", ctx, id, tx).Return(expectedErr)

		err := service.SyncCache(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockStore.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
