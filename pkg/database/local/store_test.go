package local_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/danielcbailey/Cookbook/pkg/database/local"
)

func TestNewLocal_EmptyDir(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		t.Fatal("expected non-nil database")
	}
}

func TestPersistenceAcrossRestarts(t *testing.T) {
	dir := t.TempDir()

	db, err := local.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	user := testUser()
	if err := tx.CreateUser(&user); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "store.json")); err != nil {
		t.Fatalf("store.json should exist after commit: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "store.json.tmp")); !os.IsNotExist(err) {
		t.Fatal("store.json.tmp should not exist after commit")
	}

	db2, err := local.NewLocal(dir)
	if err != nil {
		t.Fatal(err)
	}
	tx2, err := db2.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx2.Rollback()

	got, err := tx2.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("user should survive restart: %v", err)
	}
	if got.Email != user.Email {
		t.Fatalf("got email %q, want %q", got.Email, user.Email)
	}
}
