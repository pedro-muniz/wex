package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"wex/api_service/src/infra/dao"
)

func TestTransactionRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	postgresDAO := dao.NewPostgresDAO(db)
	repo := NewTransactionRepository(postgresDAO)
	ctx := context.Background()
	id := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "description", "transaction_date", "amount", "status", "created_at", "updated_at"}).
			AddRow(id, "Test", time.Now(), decimal.NewFromInt(100), "COMPLETED", time.Now(), time.Now())

		mock.ExpectQuery("SELECT id, description, transaction_date, amount, status, created_at, updated_at FROM purchase_transactions WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(rows)

		tx, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "Test", tx.Description)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).WillReturnError(sqlmock.ErrCancelled) // Simulating error or no rows
		_, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
	})
}
