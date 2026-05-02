package services

import (
	"context"

	"github.com/google/uuid"
	"wex/api_service/src/core/ports"
)

type ConversionProducerService struct {
	payloadStore ports.PayloadStore
	publisher    ports.MessagePublisher
}

func NewConversionProducerService(payloadStore ports.PayloadStore, publisher ports.MessagePublisher) *ConversionProducerService {
	return &ConversionProducerService{
		payloadStore: payloadStore,
		publisher:    publisher,
	}
}

func (s *ConversionProducerService) RequestConversion(ctx context.Context, id uuid.UUID, currency string) error {
	return s.publisher.PublishConversionRequest(ctx, id, currency)
}

func (s *ConversionProducerService) GetConversionResult(ctx context.Context, key string) (string, error) {
	return s.payloadStore.GetRaw(ctx, key)
}
