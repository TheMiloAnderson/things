package main

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey int

const userIDContextKey contextKey = 1

func userIDFromRequest(r *http.Request) int {
	v := r.Context().Value(userIDContextKey)
	uid, _ := v.(int)
	return uid
}

// Authenticator reports whether the request is logged in and the user id.
type Authenticator interface {
	AuthenticatedUserID(r *http.Request) (userID int, ok bool)
}

// SessionAuthenticator uses the gorilla session cookie (production).
type SessionAuthenticator struct {
	Store *sessions.CookieStore
}

func (s *SessionAuthenticator) AuthenticatedUserID(r *http.Request) (int, bool) {
	session, _ := s.Store.Get(r, "session-name")
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		return 0, false
	}
	uid, ok := session.Values["user_id"].(int)
	if !ok {
		return 0, false
	}
	return uid, true
}

func (a *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := a.Auth.AuthenticatedUserID(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		u, err := a.Data.GetUser(uid)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if u.PasswordChangedAt.Valid {
			session, _ := a.Store.Get(r, "session-name")
			issued, _ := session.Values["issued_at"].(int64)
			if issued == 0 || u.PasswordChangedAt.Time.Unix() > issued {
				session.Values["authenticated"] = false
				session.Options.MaxAge = -1
				_ = session.Save(r, w)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, uid)
		next(w, r.WithContext(ctx))
	}
}
