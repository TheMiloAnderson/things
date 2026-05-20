package main

import (
	"net/http"
	"sync"
	"time"
)

const (
	authLimitMax    = 5
	authLimitWindow = 15 * time.Minute
)

type authRateLimiters struct {
	Signup *loginLimiter
	Forgot *loginLimiter
	Resend *loginLimiter

	mu          sync.Mutex
	resendLast  map[string]time.Time
	forgotTimes map[string][]time.Time
}

func newAuthRateLimiters() *authRateLimiters {
	return &authRateLimiters{
		Signup:      newLoginLimiter(authLimitMax, authLimitWindow),
		Forgot:      newLoginLimiter(authLimitMax, authLimitWindow),
		Resend:      newLoginLimiter(authLimitMax, authLimitWindow),
		resendLast:  make(map[string]time.Time),
		forgotTimes: make(map[string][]time.Time),
	}
}

func (a *authRateLimiters) blockedSignup(r *http.Request) bool {
	return a.Signup.Blocked("signup:" + loginLimiterKey(r))
}

func (a *authRateLimiters) recordSignupFailure(r *http.Request) {
	a.Signup.RecordFailure("signup:" + loginLimiterKey(r))
}

func (a *authRateLimiters) blockedForgot(r *http.Request) bool {
	return a.Forgot.Blocked("forgot:" + loginLimiterKey(r))
}

func (a *authRateLimiters) recordForgotFailure(r *http.Request) {
	a.Forgot.RecordFailure("forgot:" + loginLimiterKey(r))
}

func (a *authRateLimiters) blockedResend(r *http.Request) bool {
	return a.Resend.Blocked("resend:" + loginLimiterKey(r))
}

func (a *authRateLimiters) recordResendFailure(r *http.Request) {
	a.Resend.RecordFailure("resend:" + loginLimiterKey(r))
}

// resendEmailCooldown reports whether a resend for this email was requested within 60s.
func (a *authRateLimiters) resendEmailCooldown(email string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	last, ok := a.resendLast[email]
	return ok && time.Since(last) < 60*time.Second
}

func (a *authRateLimiters) markResendEmail(email string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.resendLast[email] = time.Now()
}

// forgotEmailLimited reports whether this email has had 5+ reset requests in the past hour.
func (a *authRateLimiters) forgotEmailLimited(email string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	cutoff := time.Now().Add(-time.Hour)
	fs := a.trimForgotLocked(email, cutoff)
	return len(fs) >= 5
}

func (a *authRateLimiters) markForgotEmail(email string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	cutoff := time.Now().Add(-time.Hour)
	fs := a.trimForgotLocked(email, cutoff)
	a.forgotTimes[email] = append(fs, time.Now())
}

func (a *authRateLimiters) trimForgotLocked(email string, cutoff time.Time) []time.Time {
	fs := a.forgotTimes[email]
	j := 0
	for _, t := range fs {
		if t.After(cutoff) {
			fs[j] = t
			j++
		}
	}
	fs = fs[:j]
	if len(fs) == 0 {
		delete(a.forgotTimes, email)
		return nil
	}
	a.forgotTimes[email] = fs
	return fs
}
