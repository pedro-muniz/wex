package ports

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionResponseDTO struct {
	ID              uuid.UUID       `json:"id"`
	Description     string          `json:"description"`
	TransactionDate time.Time       `json:"transaction_date"`
	OriginalAmount  decimal.Decimal `json:"original_amount"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
