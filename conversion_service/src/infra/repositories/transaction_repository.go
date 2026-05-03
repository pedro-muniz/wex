package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/infra/queries"
)

type PostgresDAO interface {
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
}

type TransactionRepository struct {
	dao PostgresDAO
}

func NewTransactionRepository(dao PostgresDAO) *TransactionRepository {
	return &TransactionRepository{dao: dao}
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	row := r.dao.QueryRow(ctx, queries.GetTransactionByID, id)

	var tx domain.PurchaseTransaction
	err := row.Scan(&tx.ID, &tx.Description, &tx.TransactionDate, &tx.Amount, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return domain.PurchaseTransaction{}, domain.ErrNotFound
	}
	return tx, nil
}
