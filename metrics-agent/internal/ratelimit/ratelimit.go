// Package ratelimit is used to limit frequency for any operation.
package ratelimit

import (
	"context"
	"log"
	"time"
)

// TokenBucketLimiter is a channel interface for Token Bucket.
type TokenBucketLimiter struct {
	tockenBucketCh chan struct{}
}

// NewTokenBucketLimiter creates a new token bucket with "limit" capacity for "period" time period.
func NewTokenBucketLimiter(ctx context.Context, limit int, period time.Duration) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		tockenBucketCh: make(chan struct{}, limit),
	}

	for i := 0; i < limit; i++ {
		limiter.tockenBucketCh <- struct{}{}
	}

	refillInterval := period.Nanoseconds() / int64(limit)
	go limiter.startPeriodicRefill(ctx, time.Duration(refillInterval))
	return limiter
}

func (l *TokenBucketLimiter) startPeriodicRefill(ctx context.Context, interval time.Duration) {
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Ratelimiter bucket refiller was interrupted by context")
			return
		case <-timer.C:
			select {
			case l.tockenBucketCh <- struct{}{}:
			default:
			}
		}
	}
}

// Allow is a method to check for free tokens in the Token Bucket.
func (l *TokenBucketLimiter) Allow() bool {
	select {
	case <-l.tockenBucketCh:
		return true
	default:
		return false
	}
}
