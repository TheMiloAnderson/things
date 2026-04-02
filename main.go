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
	Data       HandlerData
	Auth       Authenticator
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
	nav := "templates/management_nav.html"
	pages := []string{
		"inbox.html", "login.html", "task.html", "tasks_list.html",
		"management_projects_list.html", "management_projects_new.html", "management_projects_edit.html",
		"management_areas_list.html", "management_areas_new.html",
		"management_contexts_list.html", "management_contexts_new.html",
	}
	for _, page := range pages {
		templates[page] = template.Must(template.ParseFiles(layout, nav, "templates/"+page))
	}

	return templates
}

func main() {
	dbConn := db.Connection{}
	dbConn.Connect("tasks")
	store := getStore()
	app := &App{
		Connection: &dbConn,
		Data:       &sqlHandlerData{conn: &dbConn},
		Auth:       &SessionAuthenticator{Store: store},
		Store:      store,
		Templates:  loadTemplates(),
	}
	http.HandleFunc("/login", app.loginHandler)
	http.HandleFunc("/logout", app.logoutHandler)

	http.HandleFunc("/", app.requireAuth(app.inboxHandler))
	http.HandleFunc("/tasks/", app.requireAuth(app.tasksListHandler))
	http.HandleFunc("/tasks", app.requireAuth(app.tasksListHandler))
	http.HandleFunc("/management", app.requireAuth(app.managementRedirect))
	http.HandleFunc("/management/", app.requireAuth(app.managementRoutes))
	http.HandleFunc("/api/task/", app.requireAuth(app.apiTaskHandler))
	http.HandleFunc("/api/project/", app.requireAuth(app.apiProjectHandler))
	http.HandleFunc("/api/area/", app.requireAuth(app.apiAreaHandler))
	http.HandleFunc("/api/context/", app.requireAuth(app.apiContextHandler))
	http.HandleFunc("/task/", app.requireAuth(app.taskByIDHandler))
	http.HandleFunc("/task", app.requireAuth(app.taskHandler))

	fmt.Println("Server is starting on port 8888...")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
