package apicommon

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/danielcbailey/Cookbook/internal/cache"
)

const ipRateLimit = 200 // per minute
const minWindowsBeforeExpiry = 3

func IPRateLimiter(cache cache.Cache, ip string) *RateLimiter {
	return NewRateLimiter(cache, "ip_"+ip, ipRateLimit, 1*time.Minute)
}

type RateLimiter struct {
	key      string
	interval time.Duration
	limit    int64
	cache    cache.Cache
}

func NewRateLimiter(cache cache.Cache, key string, limit int64, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		key:      key,
		interval: interval,
		limit:    limit,
		cache:    cache,
	}
}

func (r *RateLimiter) windowKey(windowIndex int64) string {
	return fmt.Sprintf("ratelimiter:%s:%d", r.key, windowIndex)
}

// Add attempts to add delta to the rate limiter's counter. Returns true if the
// request is within the limit, false if it would exceed it.
func (r *RateLimiter) Add(ctx context.Context, delta uint64) (bool, error) {
	now := time.Now()
	currentWindow := now.UnixNano() / int64(r.interval)
	elapsed := time.Duration(now.UnixNano() % int64(r.interval))

	previousCount, err := r.getWindowCount(ctx, currentWindow-1)
	if err != nil {
		return false, err
	}

	currentCount, err := r.getWindowCount(ctx, currentWindow)
	if err != nil {
		return false, err
	}

	// Sliding window estimate: weight the previous window by how much of it
	// still overlaps with the trailing interval.
	previousWeight := float64(r.interval-elapsed) / float64(r.interval)
	total := float64(currentCount) + float64(previousCount)*previousWeight

	if int64(total)+int64(delta) > r.limit {
		return false, nil
	}

	currentKey := r.windowKey(currentWindow)
	expiry := r.interval * minWindowsBeforeExpiry
	_, err = r.cache.Add(ctx, currentKey, int64(delta), expiry)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RateLimiter) getWindowCount(ctx context.Context, windowIndex int64) (int64, error) {
	val, err := r.cache.Get(ctx, r.windowKey(windowIndex))
	if err != nil {
		if r.cache.ErrIsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	if val == "" {
		return 0, nil
	}
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, nil
	}
	return count, nil
}
