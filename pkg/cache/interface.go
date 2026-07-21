package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)

	// GetInterface reads an object from the cache. Ptr should be a pointer to a json-serializable struct
	GetInterface(ctx context.Context, key string, ptr any) error

	// Set stores the value in the cache. An expiry of zero is indefinite.
	Set(ctx context.Context, key, value string, expiry time.Duration) error

	// SetInterface sets an object in the cache. Ptr should be a pointer to a json-serializable struct.
	// An expiry of zero is indefinite.
	SetInterface(ctx context.Context, key string, ptr any, expiry time.Duration) error

	Lock(ctx context.Context, key string, timeout time.Duration) error
	Unlock(ctx context.Context, key string) error

	// Add adds the delta to the key's former value. If it doesn't exist, sets to delta.
	// The expiry is updated on every call. An expiry of zero is indefinite.
	// Returns new value.
	Add(ctx context.Context, key string, delta int64, expiry time.Duration) (int64, error)

	ErrIsNotFound(err error) bool
}
