package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"things/internal/models"
	"time"
)

// safeManagementProjectReturnPath allows redirect back to a project edit page only (no open redirects).
func safeManagementProjectReturnPath(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "/management/projects/") {
		return ""
	}
	rest := strings.TrimPrefix(s, "/management/projects/")
	rest = strings.Trim(rest, "/")
	if rest == "" || strings.Contains(rest, "/") {
		return ""
	}
	id, err := strconv.Atoi(rest)
	if err != nil || id <= 0 {
		return ""
	}
	return "/management/projects/" + strconv.Itoa(id)
}

// updateProjectWithCascade persists the project; if it is done or canceled, related tasks
// that are still active or pending are set to canceled (done tasks are unchanged).
func (a *App) updateProjectWithCascade(p models.Project) error {
	if err := a.Data.UpdateProject(p); err != nil {
		return err
	}
	if p.Status == models.StatusDone || p.Status == models.StatusCanceled {
		return a.Data.CancelActiveTasksForProject(p.ID, p.UserID)
	}
	return nil
}

type InboxViewModel struct {
	Inbox    string
	Projects []models.Project
	Areas    []models.Area
}

type TaskViewModel struct {
	models.Task
	Projects         []models.Project
	Areas            []models.Area
	FormattedCreated string
}

// TasksListViewModel is the data for GET /tasks/ (open tasks: active and pending) and for the shared task_forms_list template.
type TasksListViewModel struct {
	Tasks               []models.Task
	Projects            []models.Project
	Areas               []models.Area
	ProjectPageID       int    // 0 on /tasks/; project id on management project edit (removes row if task moved off project)
	TaskFormsTitle      string // e.g. "Open tasks", "Tasks in this project"
	TaskFormsShowStatus bool   // true on project edit page: show task status column
}

func (a *App) inboxHandler(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	u, err := a.Data.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	projects, err := a.Data.AllActiveProjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	areas, err := a.Data.AllAreas(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
	}
	if r.Method == http.MethodGet {
		pageData.Data = InboxViewModel{Inbox: u.Inbox, Projects: projects, Areas: areas}
		err := a.Templates["inbox.html"].ExecuteTemplate(w, "layout.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}
		inboxText := r.FormValue("inbox")
		if err := a.Data.UpdateUserInbox(userID, inboxText); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u, err = a.Data.GetUser(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pageData.Data = InboxViewModel{Inbox: u.Inbox, Projects: projects, Areas: areas}
		err := a.Templates["inbox.html"].ExecuteTemplate(w, "layout.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (a *App) inboxProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := userIDFromRequest(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("project_name"))
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	status := models.Status(strings.TrimSpace(r.FormValue("project_status")))
	switch status {
	case models.StatusActive, models.StatusDone, models.StatusCanceled:
	default:
		status = models.StatusActive
	}
	notes := strings.TrimSpace(r.FormValue("project_notes"))
	areaID := 0
	if s := strings.TrimSpace(r.FormValue("project_area_id")); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			areaID = v
		}
	}
	p := models.Project{
		Name:   name,
		Status: status,
		Notes:  notes,
		AreaID: areaID,
		UserID: userID,
	}
	if _, err := a.Data.SaveProject(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) tasksListHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tasks" {
		http.Redirect(w, r, "/tasks/", http.StatusSeeOther)
		return
	}
	if r.URL.Path != "/tasks/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := userIDFromRequest(r)
	u, err := a.Data.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, err := a.Data.AllActiveProjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	areas, err := a.Data.AllAreas(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tasks, err := a.Data.AllActiveTasks(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
		Data: TasksListViewModel{
			Tasks:          tasks,
			Projects:       projects,
			Areas:          areas,
			TaskFormsTitle: "Open tasks",
		},
	}
	if err := a.Templates["tasks_list.html"].ExecuteTemplate(w, "layout.html", pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) taskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := userIDFromRequest(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "Task name is required", http.StatusBadRequest)
		return
	}

	status := models.Status(strings.TrimSpace(r.FormValue("status")))
	switch status {
	case models.StatusActive, models.StatusPending, models.StatusDone, models.StatusCanceled:
		// ok
	default:
		status = models.StatusActive
	}

	priorityInt, err := strconv.Atoi(strings.TrimSpace(r.FormValue("priority")))
	if err != nil {
		priorityInt = int(models.PriorityMed)
	}
	priority := models.Priority(priorityInt)
	switch priority {
	case models.PriorityLow, models.PriorityMed, models.PriorityHigh:
		// ok
	default:
		priority = models.PriorityMed
	}

	projectID := 0
	if s := strings.TrimSpace(r.FormValue("project_id")); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			projectID = v
		}
	}

	areaID := 0
	if s := strings.TrimSpace(r.FormValue("area_id")); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			areaID = v
		}
	}

	t := models.Task{}
	t.Name = name
	t.Status = status
	t.Priority = priority
	t.DateCreated = time.Now()
	t.ProjectID = projectID
	t.AreaID = areaID
	t.UserID = userID

	_, err = a.Data.SaveTask(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ret := safeManagementProjectReturnPath(r.FormValue("return_to")); ret != "" {
		http.Redirect(w, r, ret, http.StatusSeeOther)
		return
	}

	u, err := a.Data.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	projects, err := a.Data.AllActiveProjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	areas, err := a.Data.AllAreas(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
		Data: InboxViewModel{
			Inbox:    u.Inbox,
			Projects: projects,
			Areas:    areas,
		},
	}

	var buf bytes.Buffer
	if err := a.Templates["inbox.html"].ExecuteTemplate(&buf, "layout.html", pageData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(buf.Bytes())
}

func (a *App) taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Expected: /task/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/task/")
	if idStr == "" || strings.Contains(idStr, "/") {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	userID := userIDFromRequest(r)
	u, err := a.Data.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, err := a.Data.AllActiveProjects(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	areas, err := a.Data.AllAreas(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
	}

	switch r.Method {
	case http.MethodGet:
		t, err := a.Data.GetTask(id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if t.UserID != userID {
			http.NotFound(w, r)
			return
		}

		formatted := ""
		if !t.DateCreated.IsZero() {
			if loc, err := time.LoadLocation("America/Los_Angeles"); err == nil {
				formatted = t.DateCreated.In(loc).Format("January 2, 2006 at 3:04 PM MST")
			} else {
				formatted = t.DateCreated.Format("January 2, 2006 at 3:04 PM MST")
			}
		}
		vm := TaskViewModel{Task: t, Projects: projects, Areas: areas, FormattedCreated: formatted}
		pageData.Data = vm
		if err := a.Templates["task.html"].ExecuteTemplate(w, "layout.html", pageData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not parse form", http.StatusBadRequest)
			return
		}

		t, err := a.Data.GetTask(id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if t.UserID != userID {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "Task name is required", http.StatusBadRequest)
			return
		}
		t.Name = name

		status := models.Status(strings.TrimSpace(r.FormValue("status")))
		switch status {
		case models.StatusActive, models.StatusPending, models.StatusDone, models.StatusCanceled:
			t.Status = status
		}

		priorityInt, err := strconv.Atoi(strings.TrimSpace(r.FormValue("priority")))
		if err == nil {
			p := models.Priority(priorityInt)
			switch p {
			case models.PriorityLow, models.PriorityMed, models.PriorityHigh:
				t.Priority = p
			}
		}

		projectID := 0
		if s := strings.TrimSpace(r.FormValue("project_id")); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				projectID = v
			}
		}
		areaID := 0
		if s := strings.TrimSpace(r.FormValue("area_id")); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				areaID = v
			}
		}
		t.ProjectID = projectID
		t.AreaID = areaID

		if err := a.Data.UpdateTask(t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/task/"+strconv.Itoa(id), http.StatusSeeOther)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
