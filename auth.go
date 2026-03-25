package main

import (
	"net/http"
	"os"
	"things/internal/models"

	"github.com/gorilla/sessions"
	"github.com/lpernett/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func getStore() *sessions.CookieStore {
	godotenv.Load()
	session_key := os.Getenv("SESSION_KEY")
	store := sessions.NewCookieStore([]byte(session_key))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
	}
	return store
}

func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pageData := PageData{
			IsAuthenticated: false,
			Username:        "",
			Data:            nil,
		}
		err := a.Templates["login.html"].ExecuteTemplate(w, "layout.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	name := r.FormValue("username")
	pass := r.FormValue("password")
	u := models.User{Connection: *a.Connection}
	err := u.GetByName(name)
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	session, _ := a.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = u.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
