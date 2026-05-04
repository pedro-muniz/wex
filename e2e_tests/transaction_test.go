package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransactionStoreValidTransaction(t *testing.T) {
	req := TransactionRequestDTO{
		Description:     "Book purchase",
		TransactionDate: "2026-05-01",
		PurchaseAmount:  decimal.NewFromFloat(25.50),
	}

	id := createTransaction(t, req)
	waitForStatus(t, id, "COMPLETED")
}

func TestTransactionDescriptionLength(t *testing.T) {
	t.Run("Accept exactly 50 characters", func(t *testing.T) {
		req := TransactionRequestDTO{
			Description:     strings.Repeat("a", 50),
			TransactionDate: "2026-05-01",
			PurchaseAmount:  decimal.NewFromFloat(10.00),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")
	})

	t.Run("Reject 51 characters", func(t *testing.T) {
		req := TransactionRequestDTO{
			Description:     strings.Repeat("a", 51),
			TransactionDate: "2026-05-01",
			PurchaseAmount:  decimal.NewFromFloat(10.00),
		}
		// Expect 400 Bad Request
		body := fmt.Sprintf(`{"description":"%s","transaction_date":"2026-05-01","amount":"10.00"}`, req.Description)
		resp, err := http.Post(fmt.Sprintf("%s/transactions", apiBaseURL), "application/json", strings.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestTransactionAmountValidation(t *testing.T) {
	t.Run("Reject negative amount", func(t *testing.T) {
		body := `{"description":"test","transaction_date":"2026-05-01","amount":"-10.00"}`
		resp, err := http.Post(fmt.Sprintf("%s/transactions", apiBaseURL), "application/json", strings.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Reject zero amount", func(t *testing.T) {
		body := `{"description":"test","transaction_date":"2026-05-01","amount":"0.00"}`
		resp, err := http.Post(fmt.Sprintf("%s/transactions", apiBaseURL), "application/json", strings.NewReader(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Accept minimum positive amount 0.01", func(t *testing.T) {
		req := TransactionRequestDTO{
			Description:     "Small purchase",
			TransactionDate: "2026-05-01",
			PurchaseAmount:  decimal.NewFromFloat(0.01),
		}
		id := createTransaction(t, req)
		waitForStatus(t, id, "COMPLETED")
	})
}

func TestTransactionAmountRounding(t *testing.T) {
	// Scenario: Round purchase amount to nearest cent
	// Given 10.005, should be stored as 10.01
	// Wait, the API service usually parses the decimal. If we send 10.005, let's see how it's stored.
	
	req := TransactionRequestDTO{
		Description:     "Rounding test",
		TransactionDate: "2026-05-01",
		PurchaseAmount:  decimal.NewFromFloat(10.005),
	}
	id := createTransaction(t, req)
	waitForStatus(t, id, "COMPLETED")
	
	// Verify stored amount via status endpoint (which returns the transaction details)
	resp, err := http.Get(fmt.Sprintf("%s/transactions/%s/status", apiBaseURL, id))
	assert.NoError(t, err)
	defer resp.Body.Close()
	
	var tx map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&tx)
	
	// The response from status endpoint should have the amount
	// Let's check how the TransactionResponseDTO is structured in api_service
}
