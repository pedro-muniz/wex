package ports

import (
	"context"

	"github.com/google/uuid"
)

// MessagePublisher defines queue operations.
type MessagePublisher interface {
	PublishJob(ctx context.Context, jobID uuid.UUID) error
	PublishConversionRequest(ctx context.Context, jobID uuid.UUID, currency string) error
}
