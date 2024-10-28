package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (limiter *FixedWindowLimiter) Allow(ip string) (bool, time.Duration) {
	limiter.Lock()
	count, exists := limiter.clients[ip]
	limiter.Unlock()
	if !exists || count < limiter.limit {
		if !exists {
			go limiter.resetCount(ip)
		}
		limiter.Lock()
		limiter.clients[ip]++
		limiter.Unlock()
		return true, 0
	}
	return false, limiter.window
}

func (limiter *FixedWindowLimiter) resetCount(ip string) {
	time.Sleep(limiter.window)
	limiter.Lock()
	delete(limiter.clients, ip)
	limiter.Unlock()
}
