package main

import (
	"html/template"
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

func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := getStore()
		session, _ := store.Get(r, "session-name")
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles(
			"templates/layout.html",
			"templates/login.html",
		))
		tmpl.ExecuteTemplate(w, "layout.html", nil)
		return
	}
	name := r.FormValue("username")
	pass := r.FormValue("password")
	u := models.User{}
	u.Connect(dbName)
	err := u.GetByName(name)
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	store := getStore()
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = u.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	store := getStore()
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
