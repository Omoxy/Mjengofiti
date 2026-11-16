package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// RateLimiter limits requests per client IP based on token bucket algorithm.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	rate     float64
	capacity float64
}

// NewRateLimiter initializes an IP based token bucket rate limiter.
func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}
	if burst <= 0 {
		burst = 10
	}
	return &RateLimiter{
		clients:  make(map[string]*clientBucket),
		rate:     float64(requestsPerMinute) / 60.0,
		capacity: float64(burst),
	}
}

// Allow checks if the client IP is allowed to proceed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.clients[ip]
	if !exists {
		if len(rl.clients) > 10000 {
			rl.cleanup(now)
		}
		rl.clients[ip] = &clientBucket{
			tokens:     rl.capacity - 1,
			lastUpdate: now,
		}
		return true
	}

	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > rl.capacity {
		bucket.tokens = rl.capacity
	}
	bucket.lastUpdate = now

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}
	return false
}

func (rl *RateLimiter) cleanup(now time.Time) {
	for ip, b := range rl.clients {
		if now.Sub(b.lastUpdate) > 10*time.Minute {
			delete(rl.clients, ip)
		}
	}
}

// Middleware creates an HTTP middleware that limits incoming requests.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractClientIP(r.RemoteAddr)
		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limit_exceeded","message":"Too many requests. Please retry later."}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.Trim(remoteAddr, "[]")
	}
	return host
}
