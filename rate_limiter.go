package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// loginLimiter is a small in-process sliding-window rate limiter used by the
// login endpoint. It is keyed by client identifier (typically remote IP) and
// counts only failed attempts. After Max failures within Window the key is
// blocked for the remainder of the window.
type loginLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
	window   time.Duration
	max      int
	lastGC   time.Time
}

func newLoginLimiter(max int, window time.Duration) *loginLimiter {
	return &loginLimiter{
		failures: make(map[string][]time.Time),
		window:   window,
		max:      max,
		lastGC:   time.Now(),
	}
}

// Blocked reports whether the key currently exceeds the limit. Callers must
// still call RecordFailure on subsequent failed attempts.
func (l *loginLimiter) Blocked(key string) bool {
	if l == nil || l.max <= 0 {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.gcLocked()
	cutoff := time.Now().Add(-l.window)
	fs := l.trimLocked(key, cutoff)
	return len(fs) >= l.max
}

// RecordFailure appends a failure timestamp for the given key.
func (l *loginLimiter) RecordFailure(key string) {
	if l == nil || l.max <= 0 {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-l.window)
	fs := l.trimLocked(key, cutoff)
	l.failures[key] = append(fs, time.Now())
}

// Reset clears recorded failures for the key after a successful login.
func (l *loginLimiter) Reset(key string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

func (l *loginLimiter) trimLocked(key string, cutoff time.Time) []time.Time {
	fs := l.failures[key]
	j := 0
	for _, t := range fs {
		if t.After(cutoff) {
			fs[j] = t
			j++
		}
	}
	fs = fs[:j]
	if len(fs) == 0 {
		delete(l.failures, key)
		return nil
	}
	l.failures[key] = fs
	return fs
}

// gcLocked removes expired entries from the whole map at most once per window
// to keep memory bounded under high cardinality (eg, attack scenarios).
func (l *loginLimiter) gcLocked() {
	if time.Since(l.lastGC) < l.window {
		return
	}
	cutoff := time.Now().Add(-l.window)
	for k := range l.failures {
		l.trimLocked(k, cutoff)
	}
	l.lastGC = time.Now()
}

// loginLimiterKey returns the rate-limit key for a request. It prefers the
// last entry in X-Forwarded-For when the request came from a loopback
// connection (eg, the local nginx reverse proxy), otherwise it uses the
// remote address. Note: this is only meaningful if your proxy is configured
// with `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;`.
func loginLimiterKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if isLoopback(host) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			last := strings.TrimSpace(parts[len(parts)-1])
			if last != "" {
				return last
			}
		}
		if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
			return xrip
		}
	}
	return host
}

func isLoopback(host string) bool {
	if host == "" {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
