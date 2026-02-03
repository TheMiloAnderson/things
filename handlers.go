package main

import (
	"net/http"
	"things/internal/models"
)

func (a *App) inboxHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session-name")
	auth, _ := session.Values["authenticated"].(bool)
	userID := session.Values["user_id"].(int)
	u := models.User{Connection: *a.Connection}
	u.GetById(userID)
	pageData := PageData{
		IsAuthenticated: auth,
		Username:        u.Name,
	}
	if r.Method == http.MethodGet {
		pageData.Data = u.Inbox
		err := a.Templates["inbox.html"].ExecuteTemplate(w, "layout.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
		err := a.Templates["inbox.html"].ExecuteTemplate(w, "layout.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
