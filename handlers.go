package main

import (
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
	Contexts []models.Context
}

type TaskViewModel struct {
	models.Task
	Projects         []models.Project
	Areas            []models.Area
	FormattedCreated string
}

// TasksListViewModel is the data for the read-only task list (tasks page and project edit).
type TasksListViewModel struct {
	Tasks          []models.Task
	Projects       []models.Project
	Areas          []models.Area
	Contexts       []models.Context
	AreasByID      map[int]string // area ID -> name
	ProjectsByID   map[int]string // project ID -> name
	ProjectPageID  int            // 0 on /tasks/; set on project edit (for empty-state copy)
	TaskFormsTitle string         // e.g. "Open tasks", "Tasks in this project"
	FilterSort      string
	FilterStatus    string
	FilterProjectID int
	FilterAreaID    int
	FilterContextID int
	ListHeading     string // e.g. "Active Business Hours Tasks"
}

func taskListHeading(status models.Status, contextID int, contexts []models.Context) string {
	statusLabels := map[models.Status]string{
		"":                  "Active & Pending",
		models.StatusActive:   "Active",
		models.StatusPending:  "Pending",
		models.StatusDone:     "Done",
		models.StatusCanceled: "Canceled",
	}
	label := statusLabels[status]
	if label == "" {
		label = string(status)
	}
	parts := []string{label}
	if contextID > 0 {
		for _, c := range contexts {
			if c.ID == contextID {
				parts = append(parts, c.Name)
				break
			}
		}
	}
	return strings.Join(parts, " ") + " Tasks"
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
	contexts, err := a.Data.AllContextsForUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageData := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
	}
	if r.Method == http.MethodGet {
		pageData.Data = InboxViewModel{Inbox: u.Inbox, Projects: projects, Areas: areas, Contexts: contexts}
		a.render(w, r, "inbox.html", pageData)
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
		pageData.Data = InboxViewModel{Inbox: u.Inbox, Projects: projects, Areas: areas, Contexts: contexts}
		a.render(w, r, "inbox.html", pageData)
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

	sortParam := r.URL.Query().Get("sort")
	statusParam := models.Status(strings.TrimSpace(r.URL.Query().Get("status")))
	projectID, _ := strconv.Atoi(r.URL.Query().Get("project_id"))
	areaID, _ := strconv.Atoi(r.URL.Query().Get("area_id"))
	contextID, _ := strconv.Atoi(r.URL.Query().Get("context_id"))

	filter := models.TaskFilter{
		Status:    statusParam,
		ProjectID: projectID,
		AreaID:    areaID,
		ContextID: contextID,
		Sort:      sortParam,
	}
	if filter.Sort == "" {
		filter.Sort = "created_desc"
	}

	projects, err := a.Data.AllProjectsForUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	areas, err := a.Data.AllAreas(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	contexts, err := a.Data.AllContextsForUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tasks, err := a.Data.FilteredTasks(userID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	areasByID := make(map[int]string, len(areas))
	for _, ar := range areas {
		areasByID[ar.ID] = ar.Name
	}
	projectsByID := make(map[int]string, len(projects))
	for _, p := range projects {
		projectsByID[p.ID] = p.Name
	}

	pageData := PageData{
		IsAuthenticated: true,
		Username:        u.Name,
		Data: TasksListViewModel{
			Tasks:           tasks,
			Projects:        projects,
			Areas:           areas,
			Contexts:        contexts,
			AreasByID:       areasByID,
			ProjectsByID:    projectsByID,
			TaskFormsTitle:  "Open tasks",
			ListHeading:     taskListHeading(filter.Status, filter.ContextID, contexts),
			FilterSort:      filter.Sort,
			FilterStatus:    string(filter.Status),
			FilterProjectID: filter.ProjectID,
			FilterAreaID:    filter.AreaID,
			FilterContextID: filter.ContextID,
		},
	}
	a.render(w, r, "tasks_list.html", pageData)
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

	userContexts, err := a.Data.AllContextsForUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	allowedCtx := make(map[int]struct{}, len(userContexts))
	for _, c := range userContexts {
		allowedCtx[c.ID] = struct{}{}
	}
	var contextIDs []int
	for _, s := range r.Form["context_ids"] {
		id, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := allowedCtx[id]; ok {
			contextIDs = append(contextIDs, id)
		}
	}

	notes := strings.TrimSpace(r.FormValue("notes"))

	t := models.Task{}
	t.Name = name
	t.Status = status
	t.Priority = priority
	t.Notes = notes
	t.DateCreated = time.Now()
	t.ProjectID = projectID
	t.AreaID = areaID
	t.UserID = userID
	t.ContextIDs = contextIDs

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
	contexts, err := a.Data.AllContextsForUser(userID)
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
			Contexts: contexts,
		},
	}

	w.WriteHeader(http.StatusCreated)
	a.render(w, r, "inbox.html", pageData)
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
		a.render(w, r, "task.html", pageData)
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
		t.Notes = strings.TrimSpace(r.FormValue("notes"))

		if err := a.Data.UpdateTask(t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/tasks/", http.StatusSeeOther)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
