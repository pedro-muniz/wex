package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"wex/conversion_service/src/core/domain"
)

type ValkeyDAO interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Del(ctx context.Context, key string) error
	SetStatus(ctx context.Context, transactionID string, status string, expiration time.Duration) error
	GetStatus(ctx context.Context, transactionID string) (string, error)
}

type ValkeyRepository struct {
	dao ValkeyDAO
}

func NewValkeyRepository(dao ValkeyDAO) *ValkeyRepository {
	return &ValkeyRepository{dao: dao}
}

func (s *ValkeyRepository) StorePayload(ctx context.Context, jobID uuid.UUID, payload domain.PurchaseTransaction) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.dao.Set(ctx, jobID.String(), string(data), 0)
}

func (s *ValkeyRepository) GetPayload(ctx context.Context, jobID uuid.UUID) (domain.PurchaseTransaction, error) {
	data, err := s.dao.Get(ctx, jobID.String())
	if err != nil {
		return domain.PurchaseTransaction{}, err
	}
	var payload domain.PurchaseTransaction
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return domain.PurchaseTransaction{}, err
	}
	return payload, nil
}

func (s *ValkeyRepository) UpdateStatus(ctx context.Context, jobID uuid.UUID, status domain.TransactionStatus) error {
	return s.dao.SetStatus(ctx, jobID.String(), string(status), 0)
}

func (s *ValkeyRepository) GetStatus(ctx context.Context, id uuid.UUID) (domain.TransactionStatus, error) {
	val, err := s.dao.GetStatus(ctx, id.String())
	if err != nil {
		return "", err
	}
	return domain.TransactionStatus(val), nil
}

func (s *ValkeyRepository) DeletePayload(ctx context.Context, jobID uuid.UUID) error {
	return s.dao.Del(ctx, jobID.String())
}

func (s *ValkeyRepository) SetRaw(ctx context.Context, key string, data string) error {
	return s.dao.Set(ctx, key, data, 0)
}

func (s *ValkeyRepository) GetRaw(ctx context.Context, key string) (string, error) {
	return s.dao.Get(ctx, key)
}

func (s *ValkeyRepository) SetRawWithTTL(ctx context.Context, key string, data string, expiration time.Duration) error {
	return s.dao.Set(ctx, key, data, expiration)
}

func (s *ValkeyRepository) SetNXRaw(ctx context.Context, key string, data string, expiration time.Duration) (bool, error) {
	return s.dao.SetNX(ctx, key, data, expiration)
}

func (s *ValkeyRepository) DelRaw(ctx context.Context, key string) error {
	return s.dao.Del(ctx, key)
}
