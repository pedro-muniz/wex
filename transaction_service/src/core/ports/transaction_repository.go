package ports

import (
	"context"

	"github.com/google/uuid"
	"wex/transaction_service/src/core/domain"
)

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	Save(ctx context.Context, tx domain.PurchaseTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error)
	Update(ctx context.Context, tx domain.PurchaseTransaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}
