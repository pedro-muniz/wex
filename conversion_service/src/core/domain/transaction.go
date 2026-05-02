package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "PENDING"
	StatusProcessing TransactionStatus = "PROCESSING"
	StatusCompleted  TransactionStatus = "COMPLETED"
	StatusFailed     TransactionStatus = "FAILED"
)

// PurchaseTransaction represents a purchase transaction entity in the core domain.
type PurchaseTransaction struct {
	ID              uuid.UUID
	Description     string
	TransactionDate time.Time
	Amount          decimal.Decimal
	Status          TransactionStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CurrencyConversionRate represents historical exchange rate data.
type CurrencyConversionRate struct {
	TargetCurrency string
	RateDate       time.Time
	ExchangeRate   decimal.Decimal
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks if the PurchaseTransaction follows the business rules.
func (p *PurchaseTransaction) Validate() error {
	if len(p.Description) > 50 {
		return fmt.Errorf("%w: description cannot exceed 50 characters", ErrValidation)
	}

	if p.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: amount must be a positive value", ErrValidation)
	}

	if p.TransactionDate.IsZero() {
		return fmt.Errorf("%w: transaction date is required", ErrValidation)
	}

	return nil
}
