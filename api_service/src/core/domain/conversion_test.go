package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyConversionRate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rate    CurrencyConversionRate
		wantErr bool
	}{
		{
			name: "valid rate",
			rate: CurrencyConversionRate{
				TargetCurrency: "BRL",
				RateDate:       time.Now(),
				ExchangeRate:   decimal.NewFromFloat(5.123),
			},
			wantErr: false,
		},
		{
			name: "missing currency",
			rate: CurrencyConversionRate{
				ExchangeRate: decimal.NewFromFloat(5.123),
			},
			wantErr: true,
		},
		{
			name: "negative rate",
			rate: CurrencyConversionRate{
				TargetCurrency: "BRL",
				ExchangeRate:   decimal.NewFromFloat(-1.0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rate.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConversionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ConversionRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     ConversionRequest{TargetCurrency: "BRL"},
			wantErr: false,
		},
		{
			name:    "invalid length",
			req:     ConversionRequest{TargetCurrency: "BR"},
			wantErr: true,
		},
		{
			name:    "empty",
			req:     ConversionRequest{TargetCurrency: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
