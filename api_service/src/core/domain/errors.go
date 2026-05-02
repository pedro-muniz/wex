package domain

import (
	"errors"
)

var (
	ErrValidation        = errors.New("validation error")
	ErrNotFound          = errors.New("record not found")
	ErrNoConversionRate  = errors.New("no conversion rate available within 6 months")
	ErrInvalidUUID       = errors.New("invalid UUID")
)
