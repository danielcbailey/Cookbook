// Package local provides an in-memory implementation of the cache.Cache
// interface, so tests and local tooling can run without a Redis server.
package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/danielcbailey/Cookbook/pkg/cache"
)

// sweepInterval bounds how often a write walks the whole keyspace reclaiming
// expired entries. Reads evict lazily, but keys that are written and never read
// again — the rate limiter mints one per interval — would otherwise accumulate
// for the life of the process.
const sweepInterval = time.Minute

// lockRetryInterval is how long a blocked Lock waits before retrying. Redis
// polls every 50ms to amortize the round trip; there is none here.
const lockRetryInterval = 5 * time.Millisecond

type entry struct {
	value     string
	expiresAt time.Time // zero means indefinite
}

// localCache serializes every operation on a single mutex. Unlike the local
// database backend there are no transactions to serialize access, so each
// method locks for itself; the critical sections are all short enough that a
// RWMutex would buy nothing, and reads evict expired entries anyway.
type localCache struct {
	mu        sync.Mutex
	entries   map[string]entry
	locks     map[string]time.Time // key -> lock deadline; zero means indefinite
	lastSweep time.Time
}

func NewLocal() cache.Cache {
	return &localCache{
		entries:   make(map[string]entry),
		locks:     make(map[string]time.Time),
		lastSweep: time.Now(),
	}
}

// deadline converts a caller-supplied expiry into an absolute instant. A
// non-positive expiry is indefinite, represented by the zero time.
func deadline(expiry time.Duration) time.Time {
	if expiry <= 0 {
		return time.Time{}
	}
	return time.Now().Add(expiry)
}

func expired(at time.Time, now time.Time) bool {
	return !at.IsZero() && now.After(at)
}

// getLocked reads a live value, dropping the entry if it has expired. Callers
// must hold c.mu.
func (c *localCache) getLocked(key string) (string, bool) {
	e, ok := c.entries[key]
	if !ok {
		return "", false
	}
	if expired(e.expiresAt, time.Now()) {
		delete(c.entries, key)
		return "", false
	}
	return e.value, true
}

// sweepLocked reclaims expired entries and locks, at most once per
// sweepInterval. Callers must hold c.mu.
func (c *localCache) sweepLocked() {
	now := time.Now()
	if now.Sub(c.lastSweep) < sweepInterval {
		return
	}
	c.lastSweep = now

	for key, e := range c.entries {
		if expired(e.expiresAt, now) {
			delete(c.entries, key)
		}
	}
	for key, at := range c.locks {
		if expired(at, now) {
			delete(c.locks, key)
		}
	}
}

func (c *localCache) Get(_ context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, ok := c.getLocked(key)
	if !ok {
		return "", fmt.Errorf("%s: %w", key, cache.ErrNotFound)
	}
	return value, nil
}

func (c *localCache) GetInterface(ctx context.Context, key string, ptr any) error {
	value, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), ptr)
}

func (c *localCache) Set(_ context.Context, key, value string, expiry time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sweepLocked()
	c.entries[key] = entry{value: value, expiresAt: deadline(expiry)}
	return nil
}

func (c *localCache) SetInterface(ctx context.Context, key string, ptr any, expiry time.Duration) error {
	data, err := json.Marshal(ptr)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, string(data), expiry)
}

func (c *localCache) Lock(ctx context.Context, key string, timeout time.Duration) error {
	// As in the Redis backend, timeout is both how long we are willing to wait
	// and how long the lock survives once acquired.
	acquireBy := time.Now().Add(timeout)
	for {
		c.mu.Lock()
		c.sweepLocked()
		at, held := c.locks[key]
		if held && expired(at, time.Now()) {
			held = false
		}
		if !held {
			c.locks[key] = deadline(timeout)
			c.mu.Unlock()
			return nil
		}
		c.mu.Unlock()

		if time.Now().After(acquireBy) {
			return errors.New("lock acquisition timed out")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(lockRetryInterval):
		}
	}
}

// Unlock releases the lock unconditionally, without checking ownership, which
// matches the Redis backend's plain DEL.
func (c *localCache) Unlock(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.locks, key)
	return nil
}

func (c *localCache) Add(_ context.Context, key string, delta int64, expiry time.Duration) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sweepLocked()

	var current int64
	if value, ok := c.getLocked(key); ok {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			// Redis INCRBY rejects a non-integer value the same way.
			return 0, fmt.Errorf("value at %s is not an integer: %w", key, err)
		}
		current = parsed
	}

	updated := current + delta
	// The interface documents that the expiry is rewritten on every call and
	// that zero is indefinite. The Redis backend instead skips EXPIRE when the
	// expiry is zero, preserving any TTL already on the key; no caller passes a
	// zero expiry to Add, so the divergence is unobservable in practice.
	c.entries[key] = entry{value: strconv.FormatInt(updated, 10), expiresAt: deadline(expiry)}
	return updated, nil
}

func (_ *localCache) ErrIsNotFound(err error) bool {
	return errors.Is(err, cache.ErrNotFound)
}
