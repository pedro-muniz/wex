package ports

import (
	"context"

	"github.com/google/uuid"
	"wex/conversion_service/src/core/domain"
)

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error)
}
