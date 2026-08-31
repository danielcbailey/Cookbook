package objectstore

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when the requested object does not exist.
var ErrNotFound = errors.New("not found")

// ErrPresignUnsupported is returned by backends that cannot issue presigned URLs.
var ErrPresignUnsupported = errors.New("presigned URLs are not supported by this backend")

type ObjectStore interface {
	// StoreFile writes contents at path, overwriting any existing object.
	// Path components are separated by "/" regardless of host OS.
	StoreFile(ctx context.Context, path string, contents []byte) error
	// GetFile returns the object's contents, or ErrNotFound if it does not exist.
	GetFile(ctx context.Context, path string) ([]byte, error)
	// DeleteFile removes the object and returns the number of bytes deleted.
	// Deleting a nonexistent object is not an error and reports 0 bytes.
	DeleteFile(ctx context.Context, path string) (int64, error)

	// GetPresignedURL returns a time-limited URL granting read access to path.
	// Backends that cannot presign return ErrPresignUnsupported.
	GetPresignedURL(ctx context.Context, path string, validDuration time.Duration) (string, error)
}
