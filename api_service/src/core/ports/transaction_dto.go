package ports

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionRequestDTO struct {
	Description     string          `json:"description"`
	TransactionDate string          `json:"transaction_date"` // YYYY-MM-DD
	PurchaseAmount  decimal.Decimal `json:"purchase_amount"`
}

type TransactionResponseDTO struct {
	ID                uuid.UUID       `json:"id"`
	Description       string          `json:"description"`
	TransactionDate   time.Time       `json:"transactionDate"`
	PurchaseAmountUSD decimal.Decimal `json:"purchaseAmountUSD"`
	TargetCurrency    string          `json:"targetCurrency,omitempty"`
	ExchangeRate      decimal.Decimal `json:"exchangeRate,omitempty"`
	ConvertedAmount   decimal.Decimal `json:"convertedAmount,omitempty"`
	Message           string          `json:"message,omitempty"`
}
