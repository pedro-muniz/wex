package ports

import (
	"context"

	"github.com/google/uuid"
	"wex/api_service/src/core/domain"
)

// PayloadStore defines temporary storage for payloads.
type PayloadStore interface {
	StorePayload(ctx context.Context, jobID uuid.UUID, payload domain.PurchaseTransaction) error
	GetPayload(ctx context.Context, jobID uuid.UUID) (domain.PurchaseTransaction, error)
	UpdateStatus(ctx context.Context, jobID uuid.UUID, status domain.TransactionStatus) error
	GetStatus(ctx context.Context, jobID uuid.UUID) (domain.TransactionStatus, error)
	DeletePayload(ctx context.Context, jobID uuid.UUID) error
	
	// New methods for conversion results
	SetRaw(ctx context.Context, key string, data string) error
	GetRaw(ctx context.Context, key string) (string, error)
}
