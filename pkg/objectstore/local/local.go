package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielcbailey/Cookbook/pkg/objectstore"
)

type localStore struct {
	dir string
}

func NewLocal(dir string) (objectstore.ObjectStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &localStore{dir: dir}, nil
}

// resolve maps an object key onto an on-disk path, rejecting anything that would
// escape the store's directory. Keys come from request data, so this is a
// security boundary and not just a convenience.
func (s *localStore) resolve(path string) (string, error) {
	if path == "" {
		return "", errors.New("object path is empty")
	}
	if filepath.IsAbs(path) || strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("object path %q must be relative", path)
	}

	joined := filepath.Join(s.dir, filepath.FromSlash(path))
	rel, err := filepath.Rel(s.dir, joined)
	if err != nil {
		return "", fmt.Errorf("invalid object path %q: %w", path, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("object path %q escapes the store directory", path)
	}
	return joined, nil
}

func (s *localStore) StoreFile(_ context.Context, path string, contents []byte) error {
	dst, err := s.resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, contents, 0644); err != nil {
		return err
	}

	if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmp, dst)
}

func (s *localStore) GetFile(_ context.Context, path string) ([]byte, error) {
	src, err := s.resolve(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%s: %w", path, objectstore.ErrNotFound)
		}
		return nil, err
	}
	return data, nil
}

func (s *localStore) DeleteFile(_ context.Context, path string) (int64, error) {
	dst, err := s.resolve(path)
	if err != nil {
		return 0, err
	}

	// Size has to be read before the unlink, and a missing file just means
	// there is nothing to report.
	var size int64
	if info, err := os.Stat(dst); err == nil {
		size = info.Size()
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}

	if err := os.Remove(dst); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	return size, nil
}

func (_ *localStore) HostsURL(url string) bool {
	return strings.HasPrefix(url, "/")
}

func (_ *localStore) GetPresignedURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", objectstore.ErrPresignUnsupported
}
