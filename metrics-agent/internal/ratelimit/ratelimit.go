package ratelimit

import (
	"context"
	"time"
)

type TokenBucketLimiter struct {
	tockenBucketCh chan struct{}
}

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
			return
		case <-timer.C:
			select {
			case l.tockenBucketCh <- struct{}{}:
			default:
			}
		}
	}
}

func (l *TokenBucketLimiter) Allow() bool {
	select {
	case <-l.tockenBucketCh:
		return true
	default:
		return false
	}
}
