package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"wex/conversion_service/src/core/domain"
	"wex/conversion_service/src/core/ports"

	"github.com/google/uuid"
)

type TransactionQueryService struct {
	repo         ports.TransactionRepository
	rateProvider ports.ConversionRateProvider
	payloadStore ports.PayloadStore
}

func NewTransactionQueryService(repo ports.TransactionRepository, rateProvider ports.ConversionRateProvider, payloadStore ports.PayloadStore) *TransactionQueryService {
	return &TransactionQueryService{
		repo:         repo,
		rateProvider: rateProvider,
		payloadStore: payloadStore,
	}
}

func (s *TransactionQueryService) GetConvertedTransaction(ctx context.Context, id uuid.UUID, targetCurrency string) (ports.TransactionResponseDTO, error) {
	statusKey := fmt.Sprintf("conversion_status:%s:%s", id, targetCurrency)
	resultKey := fmt.Sprintf("conversion:%s:%s", id, targetCurrency)
	log.Printf("[ConversionService] Starting conversion for ID: %s to Currency: %s", id, targetCurrency)

	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[ConversionService] [ERROR] Failed to retrieve transaction %s: %v", id, err)
		
		resp := ports.TransactionResponseDTO{
			ID:      id,
			Message: fmt.Sprintf("Failed to retrieve transaction: %v", err),
			Status:  string(domain.StatusFailed),
		}
		s.saveResult(ctx, statusKey, resultKey, resp)
		return resp, err
	}

	resp := ports.TransactionResponseDTO{
		ID:                tx.ID,
		Description:       tx.Description,
		TransactionDate:   tx.TransactionDate,
		PurchaseAmountUSD: tx.Amount,
		Status:            string(domain.StatusCompleted),
		Message:           "Transaction found. No conversion requested or already in USD.",
	}

	if targetCurrency != "" && targetCurrency != "USD" {
		log.Printf("[ConversionService] Fetching exchange rate for %s on %s", targetCurrency, tx.TransactionDate.Format("2006-01-02"))
		rate, err := s.rateProvider.GetRate(ctx, targetCurrency, tx.TransactionDate)
		if err != nil {
			log.Printf("[ConversionService] [ERROR] Failed to fetch rate for %s: %v", targetCurrency, err)
			
			resp.Status = string(domain.StatusFailed)
			resp.Message = fmt.Sprintf("Failed to fetch exchange rate: %v", err)
			resp.TargetCurrency = targetCurrency
			
			s.saveResult(ctx, statusKey, resultKey, resp)
			return resp, err
		}

		convertedAmount := tx.Amount.Mul(rate.ExchangeRate).Round(2)
		resp.ConvertedAmount = convertedAmount
		resp.TargetCurrency = targetCurrency
		resp.ExchangeRate = rate.ExchangeRate
		resp.Message = fmt.Sprintf("Successfully converted %s USD to %s %s", tx.Amount.String(), convertedAmount.String(), targetCurrency)
		log.Printf("[ConversionService] Success: %s USD -> %s %s (Rate: %s)", tx.Amount, convertedAmount, targetCurrency, rate.ExchangeRate)
	}

	s.saveResult(ctx, statusKey, resultKey, resp)
	return resp, nil
}

func (s *TransactionQueryService) saveResult(ctx context.Context, statusKey, resultKey string, resp ports.TransactionResponseDTO) {
	s.payloadStore.SetRaw(ctx, statusKey, resp.Status)
	
	respData, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[ConversionService] [ERROR] Failed to marshal response: %v", err)
		return
	}
	
	if err := s.payloadStore.SetRaw(ctx, resultKey, string(respData)); err != nil {
		log.Printf("[ConversionService] [ERROR] Failed to store result in Valkey: %v", err)
	}
}
