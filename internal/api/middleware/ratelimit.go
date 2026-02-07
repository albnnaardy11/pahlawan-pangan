package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// SRE-Grade Rate Limiter
type IPLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPLimiter(r rate.Limit, b int) *IPLimiter {
	return &IPLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
		i.mu.Unlock()
	}

	return limiter
}

func LimitByIP(limiter *IPLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get IP (simulated, in prod use X-Forwarded-For)
			ip := r.RemoteAddr
			if !limiter.GetLimiter(ip).Allow() {
				http.Error(w, "⚡ Rate Limit Exceeded: Please slow down your requests (SRE Policy)", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
