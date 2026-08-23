package middleware

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"r2apimangahost/backend/internal/models"
)

type clientRecord struct {
	tokens     float64
	lastRefill time.Time
}

type IPRateLimiter struct {
	mu           sync.Mutex
	clients      map[string]*clientRecord
	rate         float64 // tokens per second
	requestsPM   int
	burst        float64 // max burst capacity
	trustProxy   bool
	cleanupEvery time.Duration // periodic cleanup of inactive IPs
}

// NewIPRateLimiter creates a Token Bucket rate limiter per IP with automatic memory cleanup
func NewIPRateLimiter(requestsPerMinute int, burst int, trustProxyHeaders bool) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:      make(map[string]*clientRecord),
		rate:         float64(requestsPerMinute) / 60.0,
		requestsPM:   requestsPerMinute,
		burst:        float64(burst),
		trustProxy:   trustProxyHeaders,
		cleanupEvery: 2 * time.Minute,
	}

	// Background goroutine to garbage collect inactive IP records to keep RAM minimal (<1MB)
	go limiter.cleanupStaleRecords()

	return limiter
}

func (l *IPRateLimiter) cleanupStaleRecords() {
	ticker := time.NewTicker(l.cleanupEvery)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, client := range l.clients {
			// If IP has been inactive for more than 5 minutes, remove from memory
			if now.Sub(client.lastRefill) > 5*time.Minute {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *IPRateLimiter) allow(ip string) (bool, int, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	client, exists := l.clients[ip]
	if !exists {
		client = &clientRecord{
			tokens:     l.burst - 1,
			lastRefill: now,
		}
		l.clients[ip] = client
		return true, int(client.tokens), 0
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(client.lastRefill).Seconds()
	client.tokens += elapsed * l.rate
	if client.tokens > l.burst {
		client.tokens = l.burst
	}
	client.lastRefill = now

	if client.tokens >= 1.0 {
		client.tokens -= 1.0
		remaining := int(client.tokens)
		return true, remaining, 0
	}

	// Calculate wait time until 1 token is available
	missing := 1.0 - client.tokens
	retryAfter := time.Duration((missing / l.rate) * float64(time.Second))
	return false, 0, retryAfter
}

// Middleware returns an HTTP middleware enforcing rate limits
func (l *IPRateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r, l.trustProxy)

		allowed, remaining, retryAfter := l.allow(ip)

		// Set standard Rate Limit headers
		burstLimit := strconv.Itoa(int(l.burst))
		w.Header().Set("RateLimit-Limit", burstLimit)
		w.Header().Set("RateLimit-Policy", fmt.Sprintf("%d;w=60;burst=%s", l.requestsPM, burstLimit))
		w.Header().Set("RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Limit", burstLimit)
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(models.APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Quá nhiều yêu cầu từ IP của bạn. Vui lòng thử lại sau %d giây.", int(retryAfter.Seconds())+1),
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractIP accurately gets client IP including Cloudflare CDN (CF-Connecting-IP) and proxies
func extractIP(r *http.Request, trustProxyHeaders bool) string {
	if !trustProxyHeaders {
		return remoteIP(r.RemoteAddr)
	}

	// 1. Cloudflare CDN Real IP
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// 2. Standard X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 3. X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 4. RemoteAddr fallback
	return remoteIP(r.RemoteAddr)
}

func remoteIP(remoteAddr string) string {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return ip
}
