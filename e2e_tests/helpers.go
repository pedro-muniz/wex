package e2e_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type TransactionRequestDTO struct {
	Description     string          `json:"description"`
	TransactionDate string          `json:"transaction_date"`
	PurchaseAmount  decimal.Decimal `json:"purchase_amount"`
}

type TransactionResponseDTO struct {
	ID                string          `json:"id"`
	Description       string          `json:"description"`
	TransactionDate   time.Time       `json:"transactionDate"`
	PurchaseAmountUSD decimal.Decimal `json:"purchaseAmountUSD"`
	Status            string          `json:"status"`
	ConvertedAmount   decimal.Decimal `json:"convertedAmount"`
	TargetCurrency    string          `json:"targetCurrency"`
	ExchangeRate      decimal.Decimal `json:"exchangeRate"`
	Message           string          `json:"message"`
}

var apiBaseURL = getEnv("API_BASE_URL", "http://localhost:8080")

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func createTransaction(t *testing.T, req TransactionRequestDTO) string {
	body, _ := json.Marshal(req)
	resp, err := http.Post(fmt.Sprintf("%s/transactions", apiBaseURL), "application/json", bytes.NewBuffer(body))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	id := result["id"]
	assert.NotEmpty(t, id)

	return id
}

func waitForStatus(t *testing.T, id string, targetStatus string) {
	maxAttempts := 20
	interval := 500 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(fmt.Sprintf("%s/transactions/%s/status", apiBaseURL, id))
		assert.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			status := fmt.Sprintf("%v", result["Status"])
			if status == targetStatus {
				return
			}
			if status == "FAILED" && targetStatus != "FAILED" {
				t.Fatalf("Transaction %s failed but expected %s", id, targetStatus)
			}
		}
		time.Sleep(interval)
	}
	t.Fatalf("Timeout waiting for transaction %s to reach status %s", id, targetStatus)
}

func requestConversion(t *testing.T, id string, currency string) {
	url := fmt.Sprintf("%s/transactions/%s/convert?currency=%s", apiBaseURL, id, url.QueryEscape(currency))
	resp, err := http.Post(url, "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func waitForConversion(t *testing.T, id string, currency string) TransactionResponseDTO {
	maxAttempts := 20
	interval := 500 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		url := fmt.Sprintf("%s/transactions/%s/convert?currency=%s", apiBaseURL, id, url.QueryEscape(currency))
		resp, err := http.Get(url)
		assert.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var result TransactionResponseDTO
			err := json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)
			return result
		}
		time.Sleep(interval)
	}
	t.Fatalf("Timeout waiting for conversion result for transaction %s in %s", id, currency)
	return TransactionResponseDTO{}
}
