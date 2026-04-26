package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"things/internal/db"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

type App struct {
	Connection   *db.Connection
	Data         HandlerData
	Auth         Authenticator
	Store        *sessions.CookieStore
	Templates    map[string]*template.Template
	LoginLimiter *loginLimiter
}

type PageData struct {
	IsAuthenticated bool
	Username        string
	Data            any
}

// templateFuncs are placeholders registered at parse time so templates can
// reference {{ csrfField }} and {{ csrfToken }}. The real implementations are
// bound per-request by App.render via a Clone()+Funcs() call.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"csrfField": func() template.HTML { return "" },
		"csrfToken": func() string { return "" },
	}
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
	taskListReadonly := "templates/task_list_readonly_snippet.html"
	for _, page := range pages {
		files := []string{layout, nav, "templates/" + page}
		if page == "tasks_list.html" || page == "management_projects_edit.html" {
			files = append(files, taskListReadonly)
		}
		templates[page] = template.Must(
			template.New("layout.html").Funcs(templateFuncs()).ParseFiles(files...),
		)
	}

	return templates
}

// csrfAuthKey returns the 32-byte authentication key used by gorilla/csrf to
// sign tokens. It prefers CSRF_KEY (32-byte secret) from the environment so
// tokens survive process restarts; falls back to a random key in development.
func csrfAuthKey() []byte {
	if v := os.Getenv("CSRF_KEY"); v != "" {
		k := []byte(v)
		if len(k) >= 32 {
			return k[:32]
		}
		log.Println("warning: CSRF_KEY is shorter than 32 bytes; generating a random key for this run (sessions will invalidate on restart)")
	}
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		log.Fatalf("csrf key: %v", err)
	}
	return k
}

// csrfSecure controls whether the CSRF cookie has the Secure flag set. Defaults
// to true; set CSRF_INSECURE=1 only for local plain-HTTP development.
func csrfSecure() bool {
	return os.Getenv("CSRF_INSECURE") != "1"
}

// csrfTrustedOrigins returns the list of origins (hostnames, no scheme) gorilla/csrf
// will trust in addition to same-origin. Configure via CSRF_TRUSTED_ORIGINS, comma-separated.
func csrfTrustedOrigins() []string {
	v := strings.TrimSpace(os.Getenv("CSRF_TRUSTED_ORIGINS"))
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func main() {
	dbConn := db.Connection{}
	dbConn.Connect("tasks")
	store := getStore()
	app := &App{
		Connection:   &dbConn,
		Data:         &sqlHandlerData{conn: &dbConn},
		Auth:         &SessionAuthenticator{Store: store},
		Store:        store,
		Templates:    loadTemplates(),
		LoginLimiter: newLoginLimiter(5, 5*time.Minute),
	}
	http.HandleFunc("/login", app.loginHandler)
	http.HandleFunc("/logout", app.logoutHandler)

	http.HandleFunc("/", app.requireAuth(app.inboxHandler))
	http.HandleFunc("/inbox/project", app.requireAuth(app.inboxProjectHandler))
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

	addr := "127.0.0.1:8888"
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		addr = v
	}

	csrfMiddleware := csrf.Protect(
		csrfAuthKey(),
		csrf.Secure(csrfSecure()),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.Path("/"),
		csrf.HttpOnly(true),
		csrf.CookieName("csrf_token"),
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.FieldName("csrf_token"),
		csrf.TrustedOrigins(csrfTrustedOrigins()),
	)

	var handler http.Handler = csrfMiddleware(http.DefaultServeMux)

	// When running in insecure (plain HTTP) mode, mark every request as
	// plaintext so gorilla/csrf's origin check compares http:// vs http://
	// instead of the default https:// vs http://.
	if !csrfSecure() {
		inner := handler
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inner.ServeHTTP(w, csrf.PlaintextHTTPRequest(r))
		})
	}

	fmt.Printf("Server is starting on %s...\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
