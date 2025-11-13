package main

import (
	"html/template"
	"log"
	"net/http"
	"things/internal/models"
)

var dbName = "tasks"

type InboxData struct {
	Text string
}

func main() {
	http.HandleFunc("/inbox", func(w http.ResponseWriter, r *http.Request) {
		u := models.User{}
		u.Connect(dbName)
		u.GetById(1)
		tmpl := template.Must(template.ParseFiles(
			"templates/layout.html",
			"templates/inbox.html",
		))
		if r.Method == http.MethodGet {
			data := InboxData{Text: u.Inbox}
			tmpl.ExecuteTemplate(w, "layout.html", data)
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
			data := InboxData{Text: u.Inbox}
			tmpl.ExecuteTemplate(w, "layout.html", data)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
