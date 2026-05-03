package queries

const (
	FindByCurrencyAndDate = `
		SELECT target_currency, rate_date, exchange_rate, created_at, updated_at 
		FROM currency_conversion_rates 
		WHERE target_currency = $1 AND rate_date = $2`

	UpsertCurrencyRate = `
		INSERT INTO currency_conversion_rates (target_currency, rate_date, exchange_rate) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (target_currency, rate_date) 
		DO UPDATE SET exchange_rate = EXCLUDED.exchange_rate, updated_at = NOW()`
)
