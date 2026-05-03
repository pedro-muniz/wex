package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/infra/dao"
)

func TestRatePostgresRepository_FindByCurrencyAndDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	postgresDAO := dao.NewPostgresDAO(db)
	repo := NewRatePostgresRepository(postgresDAO)
	ctx := context.Background()
	date := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)
	targetCurrency := "BRL"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"target_currency", "rate_date", "exchange_rate", "created_at", "updated_at"}).
			AddRow("BRL", date, decimal.NewFromFloat(5.0), time.Now(), time.Now())

		mock.ExpectQuery("SELECT target_currency, rate_date, exchange_rate, created_at, updated_at FROM currency_conversion_rates WHERE target_currency = \\$1 AND rate_date = \\$2").
			WithArgs(targetCurrency, "2023-10-27").
			WillReturnRows(rows)

		rate, err := repo.FindByCurrencyAndDate(ctx, targetCurrency, date)
		assert.NoError(t, err)
		assert.Equal(t, "BRL", rate.TargetCurrency)
		assert.Equal(t, "5", rate.ExchangeRate.String())
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT target_currency").
			WithArgs(targetCurrency, "2023-10-27").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindByCurrencyAndDate(ctx, targetCurrency, date)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFound, err)
	})
}

func TestRatePostgresRepository_Upsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()

	postgresDAO := dao.NewPostgresDAO(db)
	repo := NewRatePostgresRepository(postgresDAO)
	ctx := context.Background()
	date := time.Date(2023, 10, 27, 0, 0, 0, 0, time.UTC)

	rate := domain.CurrencyConversionRate{
		TargetCurrency: "BRL",
		RateDate:       date,
		ExchangeRate:   decimal.NewFromFloat(5.1),
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO currency_conversion_rates").
			WithArgs("BRL", "2023-10-27", decimal.NewFromFloat(5.1)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Upsert(ctx, rate)
		assert.NoError(t, err)
	})
}
