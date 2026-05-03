ALTER TABLE currency_conversion_rates ADD CONSTRAINT unique_target_currency_rate_date UNIQUE (target_currency, rate_date);
