package local_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/danielcbailey/Cookbook/pkg/cache"
	"github.com/danielcbailey/Cookbook/pkg/cache/local"
)

type session struct {
	Email string `json:"email"`
	Admin bool   `json:"admin"`
}

func TestSetAndGetRoundTrip(t *testing.T) {
	c := local.NewLocal()

	if err := c.Set(t.Context(), "greeting", "hello", 0); err != nil {
		t.Fatal(err)
	}

	got, err := c.Get(t.Context(), "greeting")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestSetOverwrites(t *testing.T) {
	c := local.NewLocal()

	if err := c.Set(t.Context(), "greeting", "first", 0); err != nil {
		t.Fatal(err)
	}
	if err := c.Set(t.Context(), "greeting", "second", 0); err != nil {
		t.Fatal(err)
	}

	got, err := c.Get(t.Context(), "greeting")
	if err != nil {
		t.Fatal(err)
	}
	if got != "second" {
		t.Fatalf("got %q, want %q", got, "second")
	}
}

func TestGetMissingReturnsNotFound(t *testing.T) {
	c := local.NewLocal()

	_, err := c.Get(t.Context(), "missing")
	if !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
	if !c.ErrIsNotFound(err) {
		t.Fatalf("ErrIsNotFound(%v) = false, want true", err)
	}
}

func TestInterfaceRoundTrip(t *testing.T) {
	c := local.NewLocal()

	want := session{Email: "test@example.com", Admin: true}
	if err := c.SetInterface(t.Context(), "session", want, 0); err != nil {
		t.Fatal(err)
	}

	var got session
	if err := c.GetInterface(t.Context(), "session", &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestGetInterfaceMissingReturnsNotFound(t *testing.T) {
	c := local.NewLocal()

	var got session
	err := c.GetInterface(t.Context(), "missing", &got)
	if !c.ErrIsNotFound(err) {
		t.Fatalf("got %v, want a not-found error", err)
	}
}

func TestValueExpires(t *testing.T) {
	c := local.NewLocal()

	if err := c.Set(t.Context(), "short", "value", 20*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(t.Context(), "short"); err != nil {
		t.Fatalf("value should still be live: %v", err)
	}

	time.Sleep(40 * time.Millisecond)

	if _, err := c.Get(t.Context(), "short"); !c.ErrIsNotFound(err) {
		t.Fatalf("got %v, want a not-found error after expiry", err)
	}
}

func TestZeroExpiryIsIndefinite(t *testing.T) {
	c := local.NewLocal()

	if err := c.Set(t.Context(), "forever", "value", 0); err != nil {
		t.Fatal(err)
	}

	time.Sleep(40 * time.Millisecond)

	got, err := c.Get(t.Context(), "forever")
	if err != nil {
		t.Fatal(err)
	}
	if got != "value" {
		t.Fatalf("got %q, want %q", got, "value")
	}
}

func TestAddCreatesAndAccumulates(t *testing.T) {
	c := local.NewLocal()

	got, err := c.Add(t.Context(), "counter", 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 5 {
		t.Fatalf("got %d, want 5", got)
	}

	got, err = c.Add(t.Context(), "counter", 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 8 {
		t.Fatalf("got %d, want 8", got)
	}

	stored, err := c.Get(t.Context(), "counter")
	if err != nil {
		t.Fatal(err)
	}
	if stored != "8" {
		t.Fatalf("got %q, want %q", stored, "8")
	}
}

func TestAddRefreshesExpiry(t *testing.T) {
	c := local.NewLocal()

	if _, err := c.Add(t.Context(), "counter", 1, 30*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := c.Add(t.Context(), "counter", 1, 30*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	// Past the first expiry but within the refreshed one.
	time.Sleep(20 * time.Millisecond)
	got, err := c.Get(t.Context(), "counter")
	if err != nil {
		t.Fatalf("expiry should have been refreshed: %v", err)
	}
	if got != "2" {
		t.Fatalf("got %q, want %q", got, "2")
	}
}

func TestAddOnNonIntegerValueErrors(t *testing.T) {
	c := local.NewLocal()

	if err := c.Set(t.Context(), "text", "not a number", 0); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Add(t.Context(), "text", 1, 0); err == nil {
		t.Fatal("expected an error adding to a non-integer value")
	}
}

func TestLockAndUnlock(t *testing.T) {
	c := local.NewLocal()

	if err := c.Lock(t.Context(), "resource", time.Second); err != nil {
		t.Fatal(err)
	}
	if err := c.Unlock(t.Context(), "resource"); err != nil {
		t.Fatal(err)
	}
	if err := c.Lock(t.Context(), "resource", time.Second); err != nil {
		t.Fatalf("lock should be free after unlock: %v", err)
	}
}

func TestLockBlocksUntilUnlocked(t *testing.T) {
	c := local.NewLocal()

	// A long TTL so the waiter can only be released by the Unlock below.
	if err := c.Lock(t.Context(), "resource", 10*time.Second); err != nil {
		t.Fatal(err)
	}

	acquired := make(chan error, 1)
	go func() {
		acquired <- c.Lock(t.Context(), "resource", 5*time.Second)
	}()

	select {
	case err := <-acquired:
		t.Fatalf("lock was acquired while held: %v", err)
	case <-time.After(30 * time.Millisecond):
	}

	if err := c.Unlock(t.Context(), "resource"); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-acquired:
		if err != nil {
			t.Fatalf("waiter failed to acquire after unlock: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter did not acquire the lock after unlock")
	}
}

func TestLockTimesOut(t *testing.T) {
	c := local.NewLocal()

	if err := c.Lock(t.Context(), "resource", 10*time.Second); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	err := c.Lock(t.Context(), "resource", 30*time.Millisecond)
	if err == nil {
		t.Fatal("expected a timeout acquiring a held lock")
	}
	if elapsed := time.Since(start); elapsed < 30*time.Millisecond {
		t.Fatalf("returned after %v, want at least the 30ms timeout", elapsed)
	}
}

func TestHeldLockExpires(t *testing.T) {
	c := local.NewLocal()

	// Never unlocked; the lock's own TTL must release it.
	if err := c.Lock(t.Context(), "resource", 20*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	time.Sleep(40 * time.Millisecond)

	if err := c.Lock(t.Context(), "resource", time.Second); err != nil {
		t.Fatalf("expired lock should be acquirable: %v", err)
	}
}

func TestLockRespectsContextCancellation(t *testing.T) {
	c := local.NewLocal()

	if err := c.Lock(t.Context(), "resource", 10*time.Second); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := c.Lock(ctx, "resource", 5*time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestConcurrentAddIsExact(t *testing.T) {
	c := local.NewLocal()

	const goroutines = 8
	const perGoroutine = 250

	wg := sync.WaitGroup{}
	for range goroutines {
		wg.Go(func() {
			for range perGoroutine {
				if _, err := c.Add(t.Context(), "counter", 1, time.Minute); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()

	got, err := c.Get(t.Context(), "counter")
	if err != nil {
		t.Fatal(err)
	}
	want := strconv.Itoa(goroutines * perGoroutine)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	c := local.NewLocal()

	const goroutines = 16
	const iterations = 100

	wg := sync.WaitGroup{}
	for i := range goroutines {
		wg.Go(func() {
			key := fmt.Sprintf("key-%d", i%4)
			for j := range iterations {
				if err := c.Set(t.Context(), key, strconv.Itoa(j), time.Minute); err != nil {
					t.Error(err)
					return
				}
				if _, err := c.Get(t.Context(), key); err != nil && !c.ErrIsNotFound(err) {
					t.Error(err)
					return
				}
				if err := c.SetInterface(t.Context(), key+":obj", session{Email: key}, time.Minute); err != nil {
					t.Error(err)
					return
				}
				var s session
				if err := c.GetInterface(t.Context(), key+":obj", &s); err != nil && !c.ErrIsNotFound(err) {
					t.Error(err)
					return
				}
				if err := c.Lock(t.Context(), key, time.Second); err == nil {
					if err := c.Unlock(t.Context(), key); err != nil {
						t.Error(err)
						return
					}
				}
			}
		})
	}
	wg.Wait()
}
