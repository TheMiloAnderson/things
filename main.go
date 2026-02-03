package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"things/internal/db"

	"github.com/gorilla/sessions"
)

type App struct {
	Connection *db.Connection
	Store      *sessions.CookieStore
	Templates  map[string]*template.Template
}

type PageData struct {
	IsAuthenticated bool
	Username        string
	Data            any
}

func loadTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	layout := "templates/layout.html"
	pages := []string{"inbox.html", "login.html"}
	for _, page := range pages {
		templates[page] = template.Must(template.ParseFiles(layout, "templates/"+page))
	}

	return templates
}

func main() {
	dbConn := db.Connection{}
	dbConn.Connect("tasks")
	store := getStore()
	app := &App{
		Connection: &dbConn,
		Store:      store,
		Templates:  loadTemplates(),
	}
	http.HandleFunc("/", AuthRequired(app.inboxHandler))

	http.HandleFunc("/login", app.loginHandler)
	http.HandleFunc("/logout", app.logoutHandler)

	fmt.Println("Server is starting on port 8888...")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
