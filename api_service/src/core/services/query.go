package services

import (
	"context"

	"github.com/google/uuid"
	"wex/api_service/src/core/ports"
)

type TransactionQueryService struct {
	repo         ports.TransactionRepository
	rateProvider ports.ConversionRateProvider
}

func NewTransactionQueryService(repo ports.TransactionRepository, rateProvider ports.ConversionRateProvider) *TransactionQueryService {
	return &TransactionQueryService{
		repo:         repo,
		rateProvider: rateProvider,
	}
}

func (s *TransactionQueryService) GetConvertedTransaction(ctx context.Context, id uuid.UUID, targetCurrency string) (ports.TransactionResponseDTO, error) {
	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ports.TransactionResponseDTO{}, err
	}

	resp := ports.TransactionResponseDTO{
		ID:                tx.ID,
		Description:       tx.Description,
		TransactionDate:   tx.TransactionDate,
		PurchaseAmountUSD: tx.Amount,
	}

	if targetCurrency != "" && targetCurrency != "USD" {
		rate, err := s.rateProvider.GetRate(ctx, targetCurrency, tx.TransactionDate)
		if err != nil {
			return ports.TransactionResponseDTO{}, err
		}

		convertedAmount := tx.Amount.Mul(rate.ExchangeRate).Round(2)
		resp.ConvertedAmount = convertedAmount
		resp.TargetCurrency = targetCurrency
		resp.ExchangeRate = rate.ExchangeRate
	}

	return resp, nil
}
