package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"wex/api_service/src/core/domain"
	"wex/api_service/src/core/ports"
)

type TransactionProducerService struct {
	payloadStore ports.PayloadStore
	publisher    ports.MessagePublisher
}

func NewTransactionProducerService(payloadStore ports.PayloadStore, publisher ports.MessagePublisher) *TransactionProducerService {
	return &TransactionProducerService{
		payloadStore: payloadStore,
		publisher:    publisher,
	}
}

func (s *TransactionProducerService) CreateTransaction(ctx context.Context, dto ports.TransactionRequestDTO) (uuid.UUID, error) {
	date, err := time.Parse("2006-01-02", dto.TransactionDate)
	if err != nil {
		return uuid.Nil, domain.ErrValidation
	}

	now := time.Now()
	tx := domain.PurchaseTransaction{
		ID:              uuid.New(),
		Description:     dto.Description,
		TransactionDate: date,
		Amount:          dto.PurchaseAmount,
		Status:          domain.StatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := tx.Validate(); err != nil {
		return uuid.Nil, err
	}

	if err := s.payloadStore.StorePayload(ctx, tx.ID, tx); err != nil {
		return uuid.Nil, err
	}

	if err := s.publisher.PublishJob(ctx, tx.ID); err != nil {
		return uuid.Nil, err
	}

	return tx.ID, nil
}

func (s *TransactionProducerService) GetTransactionStatus(ctx context.Context, id uuid.UUID) (domain.TransactionStatus, error) {
	return s.payloadStore.GetStatus(ctx, id)
}

func (s *TransactionProducerService) GetTransaction(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	return s.payloadStore.GetPayload(ctx, id)
}
