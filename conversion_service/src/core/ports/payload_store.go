package ports

import (
	"context"
)

// PayloadStore defines temporary storage for payloads.
type PayloadStore interface {
	SetRaw(ctx context.Context, key string, data string) error
	GetRaw(ctx context.Context, key string) (string, error)
}
