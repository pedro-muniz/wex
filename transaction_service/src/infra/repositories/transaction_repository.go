package repositories

import (
	"context"
	"database/sql"

	"wex/transaction_service/src/core/domain"
	"wex/transaction_service/src/infra/queries"

	"github.com/google/uuid"
)

type PostgresDAO interface {
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type TransactionRepository struct {
	dao PostgresDAO
}

func NewTransactionRepository(dao PostgresDAO) *TransactionRepository {
	return &TransactionRepository{dao: dao}
}

func (r *TransactionRepository) Save(ctx context.Context, tx domain.PurchaseTransaction) error {
	_, err := r.dao.Exec(ctx, queries.InsertTransaction, tx.ID, tx.Description, tx.TransactionDate,
		tx.Amount, tx.Status, tx.CreatedAt, tx.UpdatedAt)
	return err
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	row := r.dao.QueryRow(ctx, queries.GetTransactionByID, id)
	var tx domain.PurchaseTransaction
	err := row.Scan(&tx.ID, &tx.Description, &tx.TransactionDate, &tx.Amount, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.PurchaseTransaction{}, domain.ErrNotFound
		}
		return domain.PurchaseTransaction{}, err
	}
	return tx, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx domain.PurchaseTransaction) error {
	_, err := r.dao.Exec(ctx, queries.UpdateTransaction, tx.ID, tx.Description, tx.TransactionDate,
		tx.Amount, tx.Status, tx.UpdatedAt)
	return err
}

func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.dao.Exec(ctx, queries.DeleteTransaction, id)
	return err
}
