package local_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/danielcbailey/Cookbook/pkg/objectstore"
	"github.com/danielcbailey/Cookbook/pkg/objectstore/local"
)

func newStore(t *testing.T) (objectstore.ObjectStore, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := local.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	return store, dir
}

func TestStoreAndGetRoundTrip(t *testing.T) {
	store, _ := newStore(t)

	want := []byte("hello world")
	if err := store.StoreFile(t.Context(), "greeting.txt", want); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetFile(t.Context(), "greeting.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStoreOverwrites(t *testing.T) {
	store, _ := newStore(t)

	if err := store.StoreFile(t.Context(), "greeting.txt", []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := store.StoreFile(t.Context(), "greeting.txt", []byte("second")); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetFile(t.Context(), "greeting.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "second" {
		t.Fatalf("got %q, want %q", got, "second")
	}
}

func TestStoreCreatesNestedDirectories(t *testing.T) {
	store, dir := newStore(t)

	if err := store.StoreFile(t.Context(), "recipes/12/hero.jpg", []byte("jpeg")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "recipes", "12", "hero.jpg")); err != nil {
		t.Fatalf("expected nested file on disk: %v", err)
	}

	got, err := store.GetFile(t.Context(), "recipes/12/hero.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "jpeg" {
		t.Fatalf("got %q, want %q", got, "jpeg")
	}
}

func TestGetMissingReturnsNotFound(t *testing.T) {
	store, _ := newStore(t)

	_, err := store.GetFile(t.Context(), "missing.txt")
	if !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestDeleteRemovesFile(t *testing.T) {
	store, _ := newStore(t)

	if err := store.StoreFile(t.Context(), "greeting.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteFile(t.Context(), "greeting.txt"); err != nil {
		t.Fatal(err)
	}

	_, err := store.GetFile(t.Context(), "greeting.txt")
	if !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound after delete", err)
	}
}

func TestDeleteMissingIsNotAnError(t *testing.T) {
	store, _ := newStore(t)

	if err := store.DeleteFile(t.Context(), "missing.txt"); err != nil {
		t.Fatalf("delete of missing object: %v", err)
	}
}

func TestPathTraversalRejected(t *testing.T) {
	store, dir := newStore(t)

	// The parent of the store dir must stay untouched by every attempt below.
	parent := filepath.Dir(dir)
	before, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"", "../escape", "/abs/path", "a/../../b", "../"} {
		if err := store.StoreFile(t.Context(), path, []byte("payload")); err == nil {
			t.Fatalf("StoreFile(%q): expected error", path)
		}
		if _, err := store.GetFile(t.Context(), path); err == nil {
			t.Fatalf("GetFile(%q): expected error", path)
		}
		if err := store.DeleteFile(t.Context(), path); err == nil {
			t.Fatalf("DeleteFile(%q): expected error", path)
		}
	}

	after, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("traversal wrote outside the store dir: %d entries before, %d after", len(before), len(after))
	}
}

func TestPresignUnsupported(t *testing.T) {
	store, _ := newStore(t)

	url, err := store.GetPresignedURL(t.Context(), "greeting.txt", time.Hour)
	if !errors.Is(err, objectstore.ErrPresignUnsupported) {
		t.Fatalf("got %v, want ErrPresignUnsupported", err)
	}
	if url != "" {
		t.Fatalf("got url %q, want empty", url)
	}
}
