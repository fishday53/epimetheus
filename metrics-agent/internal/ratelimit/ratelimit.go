package ratelimit

import (
	"time"
)

type TokenBucketLimiter struct {
	tockenBucketCh chan struct{}
}

func NewTokenBucketLimiter(limit int, period time.Duration) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		tockenBucketCh: make(chan struct{}, limit),
	}

	for i := 0; i < limit; i++ {
		limiter.tockenBucketCh <- struct{}{}
	}

	refillInterval := period.Nanoseconds() / int64(limit)
	go limiter.startPeriodicRefill(time.Duration(refillInterval))
	return limiter
}

func (l *TokenBucketLimiter) startPeriodicRefill(interval time.Duration) {
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for range timer.C {
		select {
		case l.tockenBucketCh <- struct{}{}:
		default:
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
