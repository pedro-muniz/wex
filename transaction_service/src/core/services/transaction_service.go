package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"wex/transaction_service/src/core/domain"
	"wex/transaction_service/src/core/ports"
)

type TransactionPersistenceService struct {
	repo         ports.TransactionRepository
	payloadStore ports.PayloadStore
}

func NewTransactionPersistenceService(repo ports.TransactionRepository, payloadStore ports.PayloadStore) *TransactionPersistenceService {
	return &TransactionPersistenceService{
		repo:         repo,
		payloadStore: payloadStore,
	}
}

func (s *TransactionPersistenceService) ProcessTransaction(ctx context.Context, id uuid.UUID) error {
	// 1. Retrieve payload
	tx, err := s.payloadStore.GetPayload(ctx, id)
	if err != nil {
		log.Printf("failed to retrieve payload for %s: %v", id, err)
		s.payloadStore.UpdateStatus(ctx, id, domain.StatusFailed)
		return err
	}

	// 2. Update status to PROCESSING and save full payload
	tx.Status = domain.StatusProcessing
	tx.UpdatedAt = time.Now()
	if err := s.payloadStore.StorePayload(ctx, id, tx); err != nil {
		log.Printf("failed to update status to PROCESSING for %s: %v", id, err)
		return err
	}

	// 3. Persist to Postgres
	if err := s.repo.Save(ctx, tx); err != nil {
		log.Printf("failed to persist transaction %s: %v", id, err)
		tx.Status = domain.StatusFailed
		tx.UpdatedAt = time.Now()
		s.payloadStore.StorePayload(ctx, id, tx)
		return err
	}

	// 4. Update status to COMPLETED and save full payload
	tx.Status = domain.StatusCompleted
	tx.UpdatedAt = time.Now()
	if err := s.payloadStore.StorePayload(ctx, id, tx); err != nil {
		log.Printf("failed to update status to COMPLETED for %s: %v", id, err)
		return err
	}

	return nil
}

func (s *TransactionPersistenceService) GetTransaction(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	return s.repo.GetByID(ctx, id)
}
