package main

import (
	"html/template"
	"net/http"
	"things/internal/models"
)

func inboxHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	auth, _ := session.Values["authenticated"].(bool)
	userID := session.Values["user_id"].(int)
	u := models.User{}
	u.Connect(dbName)
	u.GetById(userID)
	pageData := PageData{
		IsAuthenticated: auth,
		Username:        u.Name,
	}
	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/inbox.html",
	))
	if r.Method == http.MethodGet {
		pageData.Data = u.Inbox
		tmpl.ExecuteTemplate(w, "layout.html", pageData)
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}
		inboxText := r.FormValue("inbox")
		u.Inbox = inboxText
		if err := u.Update(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		pageData.Data = u.Inbox
		tmpl.ExecuteTemplate(w, "layout.html", pageData)
	}
}
