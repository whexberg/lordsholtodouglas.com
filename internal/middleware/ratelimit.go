package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimiter holds rate-limiting state and supports graceful shutdown.
type RateLimiter struct {
	done chan struct{}
}

// Close stops the cleanup goroutine.
func (rl *RateLimiter) Close() {
	close(rl.done)
}

// RateLimit returns middleware that limits requests per IP using a token bucket.
// Each IP gets `rate` requests per second with a burst of `burst`.
// Call Close on the returned RateLimiter to stop the background cleanup goroutine.
func RateLimit(rate float64, burst int) (*RateLimiter, func(http.Handler) http.Handler) {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)
	rl := &RateLimiter{done: make(chan struct{})}

	// Clean up stale entries every minute.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for ip, v := range visitors {
					if time.Since(v.lastSeen) > 3*time.Minute {
						delete(visitors, ip)
					}
				}
				mu.Unlock()
			case <-rl.done:
				return
			}
		}
	}()

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only rate-limit state-changing requests.
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}

			mu.Lock()
			v, ok := visitors[ip]
			if !ok {
				v = &visitor{tokens: float64(burst)}
				visitors[ip] = v
			}

			// Refill tokens based on elapsed time.
			elapsed := time.Since(v.lastSeen).Seconds()
			v.tokens += elapsed * rate
			if v.tokens > float64(burst) {
				v.tokens = float64(burst)
			}
			v.lastSeen = time.Now()

			if v.tokens < 1 {
				mu.Unlock()
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			v.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}

	return rl, mw
}
