package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"wex/transaction_service/src/core/domain"
)

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

func TestValkeyPayloadStore(t *testing.T) {
	mockDAO := new(MockValkeyDAO)
	store := NewValkeyPayloadStore(mockDAO)
	ctx := context.Background()
	id := uuid.New()
	tx := domain.PurchaseTransaction{ID: id, Description: "Test", Amount: decimal.NewFromInt(100)}

	t.Run("StorePayload", func(t *testing.T) {
		data, _ := json.Marshal(tx)
		mockDAO.On("Set", ctx, id.String(), string(data), time.Duration(0)).Return(nil)
		err := store.StorePayload(ctx, id, tx)
		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("GetPayload", func(t *testing.T) {
		data, _ := json.Marshal(tx)
		mockDAO.On("Get", ctx, id.String()).Return(string(data), nil)
		res, err := store.GetPayload(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, tx.Description, res.Description)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		mockDAO.On("SetStatus", ctx, id.String(), string(domain.StatusCompleted), time.Duration(0)).Return(nil)
		err := store.UpdateStatus(ctx, id, domain.StatusCompleted)
		assert.NoError(t, err)
	})
}
