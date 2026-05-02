package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"wex/transaction_service/src/core/domain"
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
	const query = `
		INSERT INTO purchase_transactions (id, description, transaction_date, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.dao.Exec(ctx, query, tx.ID, tx.Description, tx.TransactionDate, tx.Amount, tx.Status, tx.CreatedAt, tx.UpdatedAt)
	return err
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.PurchaseTransaction, error) {
	const query = `
		SELECT id, description, transaction_date, amount, status, created_at, updated_at 
		FROM purchase_transactions 
		WHERE id = $1`
	row := r.dao.QueryRow(ctx, query, id)
	var tx domain.PurchaseTransaction
	err := row.Scan(&tx.ID, &tx.Description, &tx.TransactionDate, &tx.Amount, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return domain.PurchaseTransaction{}, domain.ErrNotFound
	}
	return tx, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx domain.PurchaseTransaction) error {
	const query = `
		UPDATE purchase_transactions 
		SET description = $2, transaction_date = $3, amount = $4, status = $5, updated_at = $6 
		WHERE id = $1`
	_, err := r.dao.Exec(ctx, query, tx.ID, tx.Description, tx.TransactionDate, tx.Amount, tx.Status, tx.UpdatedAt)
	return err
}

func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM purchase_transactions WHERE id = $1`
	_, err := r.dao.Exec(ctx, query, id)
	return err
}
