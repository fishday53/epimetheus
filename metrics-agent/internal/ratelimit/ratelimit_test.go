package ratelimit

import (
	"context"
	"testing"
	"time"
)

func Test_NewTokenBucketLimiter(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		period time.Duration
		wanted int
	}{
		{
			name:   "Check TokenBucketLimiter",
			limit:  10,
			period: time.Second * 1,
			wanted: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token int
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			rateLimit := NewTokenBucketLimiter(ctx, tt.limit, tt.period)

			for rateLimit.Allow() {
				token++
			}
			cancel()

			if token != tt.wanted {
				t.Errorf("RateLimiter failed: got %d tokens instead of %d", token, tt.wanted)
			}
		})
	}
}
