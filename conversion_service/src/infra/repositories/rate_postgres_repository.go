package repositories

import (
	"context"
	"database/sql"
	"time"

	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/infra/queries"
)

type RatePostgresDAO interface {
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type RatePostgresRepository struct {
	dao RatePostgresDAO
}

func NewRatePostgresRepository(dao RatePostgresDAO) *RatePostgresRepository {
	return &RatePostgresRepository{dao: dao}
}

func (r *RatePostgresRepository) FindByCurrencyAndDate(ctx context.Context, targetCurrency string, rateDate time.Time) (domain.CurrencyConversionRate, error) {
	row := r.dao.QueryRow(ctx, queries.FindByCurrencyAndDate, targetCurrency, rateDate.Format("2006-01-02"))

	var rate domain.CurrencyConversionRate
	err := row.Scan(&rate.TargetCurrency, &rate.RateDate, &rate.ExchangeRate, &rate.CreatedAt, &rate.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.CurrencyConversionRate{}, domain.ErrNotFound
		}
		return domain.CurrencyConversionRate{}, err
	}
	return rate, nil
}

func (r *RatePostgresRepository) Upsert(ctx context.Context, rate domain.CurrencyConversionRate) error {
	_, err := r.dao.Exec(ctx, queries.UpsertCurrencyRate, rate.TargetCurrency, rate.RateDate.Format("2006-01-02"), rate.ExchangeRate)
	return err
}
