package queries

const (
	FindByCurrencyAndDate = `
		SELECT target_currency, rate_date, exchange_rate, created_at, updated_at 
		FROM currency_conversion_rates 
		WHERE target_currency = $1 
		  AND rate_date <= $2::timestamp 
		  AND rate_date >= ($2::timestamp - interval '6 months')
		ORDER BY rate_date DESC 
		LIMIT 1`

	UpsertCurrencyRate = `
		INSERT INTO currency_conversion_rates (target_currency, rate_date, exchange_rate) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (target_currency, rate_date) 
		DO UPDATE SET exchange_rate = EXCLUDED.exchange_rate, updated_at = NOW()`
)
