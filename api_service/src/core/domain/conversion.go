package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// CurrencyConversionRate represents historical exchange rate data.
type CurrencyConversionRate struct {
	TargetCurrency string
	RateDate       time.Time
	ExchangeRate   decimal.Decimal
}

// Validate checks if the conversion rate data is valid.
func (c *CurrencyConversionRate) Validate() error {
	if len(c.TargetCurrency) == 0 {
		return fmt.Errorf("%w: currency code is required", ErrValidation)
	}
	if c.ExchangeRate.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: exchange rate must be positive", ErrValidation)
	}
	return nil
}

// ConversionRequest represents a request for currency conversion.
type ConversionRequest struct {
	TargetCurrency string
}

func (r *ConversionRequest) Validate() error {
	if len(r.TargetCurrency) != 3 {
		return fmt.Errorf("%w: currency code must be exactly 3 characters", ErrValidation)
	}
	return nil
}
