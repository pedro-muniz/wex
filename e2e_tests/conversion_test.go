package e2e_tests

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestConversionFlow(t *testing.T) {
	// Cleanup rates before and after tests
	CleanupRates()
	defer CleanupRates()

	t.Run("Successful conversion with exact date", func(t *testing.T) {
		date := "2026-05-01"
		currency := "Euro Zone-Euro"
		rate := 0.92
		SeedRate(currency, date, rate)

		req := TransactionRequestDTO{
			Description:     "Euro trip",
			TransactionDate: date,
			PurchaseAmount:  decimal.NewFromFloat(100.00),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")

		requestConversion(t, id, currency)
		result := waitForConversion(t, id, currency)

		assert.Equal(t, "COMPLETED", result.Status)
		assert.Equal(t, currency, result.TargetCurrency)
		assert.True(t, decimal.NewFromFloat(rate).Equal(result.ExchangeRate))
		assert.True(t, decimal.NewFromFloat(92.00).Equal(result.ConvertedAmount))
	})

	t.Run("Use closest exchange rate within 6 months", func(t *testing.T) {
		// Transaction date: 2026-05-01
		// Rate date: 2026-01-01 (4 months prior)
		txDate := "2026-05-01"
		rateDate := "2026-01-01"
		currency := "Afghanistan-Afghani"
		rate := 0.79
		SeedRate(currency, rateDate, rate)

		req := TransactionRequestDTO{
			Description:     "London calling",
			TransactionDate: txDate,
			PurchaseAmount:  decimal.NewFromFloat(100.00),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")

		requestConversion(t, id, currency)
		result := waitForConversion(t, id, currency)

		assert.Equal(t, "COMPLETED", result.Status)
		assert.True(t, decimal.NewFromFloat(rate).Equal(result.ExchangeRate))
		assert.True(t, decimal.NewFromFloat(79.00).Equal(result.ConvertedAmount))
	})

	t.Run("Use exchange rate exactly on 6-month boundary", func(t *testing.T) {
		// Transaction date: 2026-05-01
		// Rate date: 2025-11-01 (exactly 6 months prior)
		txTime, _ := time.Parse("2006-01-02", "2026-05-01")
		rateDate := txTime.AddDate(0, -6, 0).Format("2006-01-02")
		currency := "Brazil-Real"
		rate := 150.0
		SeedRate(currency, rateDate, rate)

		req := TransactionRequestDTO{
			Description:     "Tokyo drift",
			TransactionDate: txTime.Format("2006-01-02"),
			PurchaseAmount:  decimal.NewFromFloat(10.00),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")

		requestConversion(t, id, currency)
		result := waitForConversion(t, id, currency)

		assert.Equal(t, "COMPLETED", result.Status)
		assert.True(t, decimal.NewFromFloat(rate).Equal(result.ExchangeRate))
	})

	t.Run("Fail when no exchange rate is available within 6 months", func(t *testing.T) {
		// Transaction date: 2026-05-01
		// Rate date: 2025-10-25 (more than 6 months prior)
		txDate := "2026-05-01"
		rateDate := "2025-10-25"
		currency := "Non-Existent-Currency"
		rate := 1.35
		SeedRate(currency, rateDate, rate)

		req := TransactionRequestDTO{
			Description:     "Maple leaf",
			TransactionDate: txDate,
			PurchaseAmount:  decimal.NewFromFloat(100.00),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")

		requestConversion(t, id, currency)
		// The conversion service should fail. Result status should be FAILED.
		result := waitForConversion(t, id, currency)
		assert.Equal(t, "FAILED", result.Status)
		assert.Contains(t, result.Message, "exchange rate")
	})

	t.Run("Rounding converted amount to two decimals", func(t *testing.T) {
		// 100.00 * 0.92345 = 92.345 -> 92.35
		date := "2026-05-01"
		currency := "Switzerland-Franc"
		rate := 0.92345
		SeedRate(currency, date, rate)

		req := TransactionRequestDTO{
			Description:     "Swiss watch",
			TransactionDate: date,
			PurchaseAmount:  decimal.NewFromFloat(100.00),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")

		requestConversion(t, id, currency)
		result := waitForConversion(t, id, currency)

		assert.Equal(t, "COMPLETED", result.Status)
		// Expected: 92.35 (Nearest cent)
		assert.True(t, decimal.NewFromFloat(92.35).Equal(result.ConvertedAmount))
	})
}
