package encblob

import (
	"context"
	"fmt"
	"time"
)

// FallbackStorage tries primary first; on failure uses secondary.
// On primary success it fires an async secondary pin for redundancy.
type FallbackStorage struct {
	primary   Storage
	secondary Storage
}

// NewFallbackStorage builds the production two-provider storage from env vars.
// Returns ErrStorageNotConfigured only when both providers are unavailable.
func NewFallbackStorage() (*FallbackStorage, error) {
	pinata, err1 := NewPinataStorage()
	storj, err2 := NewStorjStorage()
	switch {
	case err1 != nil && err2 != nil:
		return nil, ErrStorageNotConfigured
	case err1 != nil:
		return &FallbackStorage{primary: storj, secondary: NewStubStorage()}, nil
	case err2 != nil:
		return &FallbackStorage{primary: pinata, secondary: NewStubStorage()}, nil
	default:
		return &FallbackStorage{primary: pinata, secondary: storj}, nil
	}
}

// NewFallbackStorageWithProviders wires explicit primary/secondary (tests).
func NewFallbackStorageWithProviders(primary, secondary Storage) *FallbackStorage {
	return &FallbackStorage{primary: primary, secondary: secondary}
}

// Store tries primary; falls back to secondary on error.
func (f *FallbackStorage) Store(ctx context.Context, encrypted []byte) (string, error) {
	uri, err := f.primary.Store(ctx, encrypted)
	if err == nil {
		encCopy := make([]byte, len(encrypted))
		copy(encCopy, encrypted)
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			_, _ = f.secondary.Store(bgCtx, encCopy)
		}()
		return uri, nil
	}

	uri, fallbackErr := f.secondary.Store(ctx, encrypted)
	if fallbackErr != nil {
		return "", fmt.Errorf("all storage providers failed: primary=%v fallback=%v", err, fallbackErr)
	}
	return uri, nil
}
