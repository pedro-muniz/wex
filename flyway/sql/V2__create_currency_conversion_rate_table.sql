CREATE TABLE currency_conversion_rates (
    id SERIAL PRIMARY KEY,
    target_currency VARCHAR(50) NOT NULL,
    rate_date TIMESTAMP NOT NULL,
    exchange_rate DECIMAL(19, 6) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_currency_conversion_rates_lookup ON currency_conversion_rates (target_currency, rate_date DESC);
