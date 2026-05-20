package main

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
	"things/internal/models"

	"github.com/gorilla/sessions"
	"github.com/lpernett/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// dummyBcryptHash is a valid bcrypt hash used to equalize the timing of failed
// login attempts when the user does not exist. The plaintext is never used.
const dummyBcryptHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

func getStore() *sessions.CookieStore {
	godotenv.Load()
	session_key := os.Getenv("SESSION_KEY")
	store := sessions.NewCookieStore([]byte(session_key))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   os.Getenv("CSRF_INSECURE") != "1",
		SameSite: http.SameSiteLaxMode,
	}
	return store
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		vm := AuthViewModel{}
		switch r.URL.Query().Get("verified") {
		case "1":
			vm.Flash = "verified"
		}
		switch r.URL.Query().Get("reset") {
		case "1":
			vm.Flash = "reset"
		}
		a.render(w, r, "login.html", PageData{
			IsAuthenticated: false,
			Data:            vm,
		})
		return
	}

	if a.LoginLimiter != nil {
		if a.LoginLimiter.Blocked(loginLimiterKey(r)) {
			http.Error(w, "Too many login attempts. Try again in a few minutes.", http.StatusTooManyRequests)
			return
		}
	}

	name := strings.TrimSpace(r.FormValue("username"))
	pass := r.FormValue("password")

	u := models.User{Connection: *a.Connection}
	err := u.GetByName(name)
	if err != nil {
		// Run a dummy bcrypt comparison so timing is comparable to a real check.
		_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(pass))
		if a.LoginLimiter != nil {
			a.LoginLimiter.RecordFailure(loginLimiterKey(r))
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)); err != nil {
		if a.LoginLimiter != nil {
			a.LoginLimiter.RecordFailure(loginLimiterKey(r))
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !u.IsEmailVerified() {
		if a.LoginLimiter != nil {
			a.LoginLimiter.RecordFailure(loginLimiterKey(r))
		}
		a.render(w, r, "login.html", PageData{
			IsAuthenticated: false,
			Data: AuthViewModel{
				Error:      "Please verify your email before signing in.",
				Email:      u.Email,
				ShowResend: true,
			},
		})
		return
	}

	if a.LoginLimiter != nil {
		a.LoginLimiter.Reset(loginLimiterKey(r))
	}

	session, _ := a.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = u.ID
	session.Values["issued_at"] = time.Now().Unix()
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
