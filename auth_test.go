package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"things/internal/mail"
	"things/internal/models"
)

type authFakeData struct {
	fakeHandlerData
	usersByEmail map[string]models.User
	usersByName  map[string]models.User
	created      []models.User
	tokens       map[string]authTokenRecord
	nextID       int
}

type authTokenRecord struct {
	userID    int
	purpose   models.TokenPurpose
	expiresAt time.Time
	used      bool
}

func newAuthFakeData() *authFakeData {
	return &authFakeData{
		usersByEmail: make(map[string]models.User),
		usersByName:  make(map[string]models.User),
		tokens:       make(map[string]authTokenRecord),
		nextID:       100,
	}
}

func (a *authFakeData) GetUserByEmail(email string) (models.User, error) {
	u, ok := a.usersByEmail[email]
	if !ok {
		return models.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (a *authFakeData) GetUserByName(name string) (models.User, error) {
	u, ok := a.usersByName[name]
	if !ok {
		return models.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (a *authFakeData) CreateUser(u models.User) (int64, error) {
	a.nextID++
	u.ID = a.nextID
	a.usersByEmail[u.Email] = u
	a.usersByName[u.Name] = u
	a.created = append(a.created, u)
	return int64(u.ID), nil
}

func (a *authFakeData) MarkEmailVerified(userID int) error {
	for email, u := range a.usersByEmail {
		if u.ID == userID {
			u.EmailVerifiedAt = sql.NullTime{Time: time.Now(), Valid: true}
			a.usersByEmail[email] = u
			if n, ok := a.usersByName[u.Name]; ok {
				n.EmailVerifiedAt = u.EmailVerifiedAt
				a.usersByName[u.Name] = n
			}
			return nil
		}
	}
	return sql.ErrNoRows
}

func (a *authFakeData) UpdatePassword(userID int, passwordHash string) error {
	for email, u := range a.usersByEmail {
		if u.ID == userID {
			u.PasswordHash = passwordHash
			u.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: true}
			a.usersByEmail[email] = u
			return nil
		}
	}
	return sql.ErrNoRows
}

func (a *authFakeData) IssueAuthToken(userID int, purpose models.TokenPurpose, ttl time.Duration) (string, error) {
	token := "test-token-" + string(purpose)
	a.tokens[token] = authTokenRecord{
		userID:    userID,
		purpose:   purpose,
		expiresAt: time.Now().Add(ttl),
	}
	return token, nil
}

func (a *authFakeData) ConsumeAuthToken(plaintext string, purpose models.TokenPurpose) (int, error) {
	rec, ok := a.tokens[plaintext]
	if !ok || rec.purpose != purpose {
		return 0, models.ErrTokenInvalid
	}
	if rec.used {
		return 0, models.ErrTokenUsed
	}
	if time.Now().After(rec.expiresAt) {
		return 0, models.ErrTokenExpired
	}
	rec.used = true
	a.tokens[plaintext] = rec
	return rec.userID, nil
}

func testAuthApp(data HandlerData) (*App, *mail.NoopSender) {
	sender := &mail.NoopSender{}
	app := testApp(data, &fakeAuth{OK: false})
	app.Mailer = sender
	app.BaseURL = "http://test.example"
	app.AuthLimiter = newAuthRateLimiters()
	return app, sender
}

func TestSignupHandler_SendsVerification(t *testing.T) {
	data := newAuthFakeData()
	app, sender := testAuthApp(data)

	body := url.Values{
		"username":          {"newuser"},
		"email":             {"new@example.com"},
		"password":          {"longpassword12"},
		"password_confirm":  {"longpassword12"},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.signupHandler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if len(data.created) != 1 {
		t.Fatalf("created %d users", len(data.created))
	}
	if sender.Last.To != "new@example.com" {
		t.Fatalf("email to %q", sender.Last.To)
	}
	if !strings.Contains(sender.Last.TextBody, "verify-email") {
		t.Fatalf("body should contain verify link: %s", sender.Last.TextBody)
	}
}

func TestSignupHandler_DuplicateEmail(t *testing.T) {
	data := newAuthFakeData()
	data.usersByEmail["taken@example.com"] = models.User{ID: 1, Email: "taken@example.com", Name: "x"}
	app, _ := testAuthApp(data)

	body := url.Values{
		"username":         {"other"},
		"email":            {"taken@example.com"},
		"password":         {"longpassword12"},
		"password_confirm": {"longpassword12"},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.signupHandler(w, r)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d", w.Code)
	}
}

func TestForgotPasswordHandler_SameResponse(t *testing.T) {
	data := newAuthFakeData()
	data.usersByEmail["real@example.com"] = models.User{
		ID: 1, Email: "real@example.com", Name: "Real",
		EmailVerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}
	app, sender := testAuthApp(data)

	for _, email := range []string{"real@example.com", "nobody@example.com"} {
		body := url.Values{"email": {email}}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(body.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		app.forgotPasswordHandler(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d for %s", w.Code, email)
		}
		if !strings.Contains(w.Body.String(), "If an account exists") {
			t.Fatalf("expected generic message for %s", email)
		}
	}
	if sender.Last.To != "real@example.com" {
		t.Fatalf("reset only sent for real user, got %q", sender.Last.To)
	}
}

func TestVerifyEmailHandler_Redirects(t *testing.T) {
	data := newAuthFakeData()
	data.usersByEmail["u@example.com"] = models.User{ID: 5, Email: "u@example.com", Name: "U"}
	data.tokens["tok"] = authTokenRecord{userID: 5, purpose: models.PurposeVerify, expiresAt: time.Now().Add(time.Hour)}

	app, _ := testAuthApp(data)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/verify-email?token=tok", nil)
	app.verifyEmailHandler(w, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status %d", w.Code)
	}
	u := data.usersByEmail["u@example.com"]
	if !u.IsEmailVerified() {
		t.Fatal("user should be verified")
	}
}

func TestAuthRateLimiter_SignupBlocked(t *testing.T) {
	lim := newAuthRateLimiters()
	r := httptest.NewRequest(http.MethodPost, "/signup", nil)
	r.RemoteAddr = "203.0.113.1:1234"
	key := "signup:" + loginLimiterKey(r)
	for i := 0; i < authLimitMax; i++ {
		lim.Signup.RecordFailure(key)
	}
	if !lim.blockedSignup(r) {
		t.Fatal("expected signup blocked")
	}
}

var _ HandlerData = (*authFakeData)(nil)
