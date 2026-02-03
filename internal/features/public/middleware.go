package public

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SetupRedirectDependencies interface {
	HasUser(ctx context.Context) (bool, error)
}

// WithSetupRedirect wraps the handler with additional behavior.
func WithSetupRedirect(deps SetupRedirectDependencies, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/setup" || strings.HasPrefix(r.URL.Path, "/invite/") || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		if ok, err := deps.HasUser(r.Context()); err != nil {
			http.Error(w, "Failed to load profile", http.StatusInternalServerError)
			return
		} else if !ok {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var privateRateLimiter = newRateLimiter(120, time.Minute)

// WithPrivateRateLimit throttles access to private identity routes.
func WithPrivateRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/p/") {
			next.ServeHTTP(w, r)
			return
		}
		key := clientIP(r)
		allowed, reset := privateRateLimiter.Allow(key, time.Now())
		if !allowed {
			retryAfter := int(time.Until(reset).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type rateLimiter struct {
	mu     sync.Mutex
	hits   map[string]rateBucket
	limit  int
	window time.Duration
}

type rateBucket struct {
	count int
	reset time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		hits:   make(map[string]rateBucket),
		limit:  limit,
		window: window,
	}
}

func (l *rateLimiter) Allow(key string, now time.Time) (bool, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.hits[key]
	if !ok || now.After(bucket.reset) {
		bucket = rateBucket{count: 0, reset: now.Add(l.window)}
	}
	if bucket.count >= l.limit {
		l.hits[key] = bucket
		return false, bucket.reset
	}
	bucket.count++
	l.hits[key] = bucket
	return true, bucket.reset
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if strings.TrimSpace(r.RemoteAddr) != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
