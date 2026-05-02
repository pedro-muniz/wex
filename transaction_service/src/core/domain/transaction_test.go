package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPurchaseTransaction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      PurchaseTransaction
		wantErr bool
	}{
		{
			name: "valid transaction",
			tx: PurchaseTransaction{
				Description:     "Test Purchase",
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(100.50),
			},
			wantErr: false,
		},
		{
			name: "description too long",
			tx: PurchaseTransaction{
				Description:     "This description is definitely way too long and should fail validation because it exceeds fifty characters",
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(100.50),
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			tx: PurchaseTransaction{
				Description:     "Test Purchase",
				TransactionDate: time.Now(),
				Amount:          decimal.Zero,
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			tx: PurchaseTransaction{
				Description:     "Test Purchase",
				TransactionDate: time.Now(),
				Amount:          decimal.NewFromFloat(-10.0),
			},
			wantErr: true,
		},
		{
			name: "missing date",
			tx: PurchaseTransaction{
				Description: "Test Purchase",
				Amount:      decimal.NewFromFloat(100.50),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
