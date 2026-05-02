package ports

import (
	"context"
	"time"

	"wex/api_service/src/core/domain"
)

// ConversionRateProvider defines rate fetching operations.
type ConversionRateProvider interface {
	GetRate(ctx context.Context, targetCurrency string, transactionDate time.Time) (domain.CurrencyConversionRate, error)
}
