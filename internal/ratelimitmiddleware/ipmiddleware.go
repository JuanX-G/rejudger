package ratelimitmiddleware

import (
	"net/http"
	"revit/internal/nethelpers"
	"revit/internal/ratelimiter"
	"strconv"
	"time"
)

type IpRateLimitMiddleware struct {
	lim         *ratelimiter.Limiter[string]
	waitTimeStr string
}

// Deault config for fetch or update-like endpoint.
var DefaultRateLimiterConfig = ratelimiter.RateLimiterConfig{
	TimeToEvict: time.Duration(3 * time.Second),
	CleanupTime: time.Duration(1 * time.Minute),
	MaxSize:     50,
	MaxCounter:  20,
}

func NewIpRateLimitMiddleware(cfg ratelimiter.RateLimiterConfig) *IpRateLimitMiddleware {
	return &IpRateLimitMiddleware{
		lim:         ratelimiter.NewLimiter[string](cfg),
		waitTimeStr: strconv.Itoa(int(cfg.TimeToEvict / time.Second)),
	}
}

// Returns a ratelimiter middleware as a http.HandleFunc
func (rl *IpRateLimitMiddleware) GetRateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := nethelpers.GetClientIP(r)
		if ip == "" {
			next(w, r)
		} else {
			isAllowed := rl.lim.Allow(ip)
			if !isAllowed {
				w.Header().Add("Retry-After", rl.waitTimeStr)
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				next(w, r)
			}
			return
		}
	}

}
