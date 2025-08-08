package accrualclient

import (
	"sync"
	"time"
)

type rateLimiter struct {
	mu      sync.Mutex
	blocked bool
	until   time.Time
}

func (rl *rateLimiter) isBlocked() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.blocked && time.Now().Before(rl.until)
}

func (rl *rateLimiter) blockFor(duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.blocked = true
	rl.until = time.Now().Add(duration)
}

func NewRateLimiter() *rateLimiter {
	return &rateLimiter{}
}
