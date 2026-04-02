package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"things/internal/models"
)

type fakeAuth struct {
	UID int
	OK  bool
}

func (f *fakeAuth) AuthenticatedUserID(r *http.Request) (int, bool) {
	return f.UID, f.OK
}

type fakeHandlerData struct {
	User           models.User
	GetUserErr     error
	UpdateInboxErr error
	Projects       []models.Project
	ProjectsErr    error
	Areas          []models.Area
	AreasErr       error
	Tasks          []models.Task
	TasksErr       error
	TasksByID      map[int]models.Task
	GetTaskErr     error
	SaveTaskID     int64
	SaveTaskErr    error
	UpdateErr      error

	LastUpdateTask models.Task
	UpdateCount    int

	ProjectsByID map[int]models.Project
	AreasByID    map[int]models.Area
	Contexts     []models.Context
	ContextsByID map[int]models.Context
}

func (f *fakeHandlerData) GetUser(userID int) (models.User, error) {
	if f.GetUserErr != nil {
		return models.User{}, f.GetUserErr
	}
	_ = userID
	return f.User, nil
}

func (f *fakeHandlerData) UpdateUserInbox(userID int, inbox string) error {
	if f.UpdateInboxErr != nil {
		return f.UpdateInboxErr
	}
	_ = userID
	f.User.Inbox = inbox
	return nil
}

func (f *fakeHandlerData) AllActiveProjects(userID int) ([]models.Project, error) {
	if f.ProjectsErr != nil {
		return nil, f.ProjectsErr
	}
	_ = userID
	return f.Projects, nil
}

func (f *fakeHandlerData) AllAreas(userID int) ([]models.Area, error) {
	if f.AreasErr != nil {
		return nil, f.AreasErr
	}
	_ = userID
	return f.Areas, nil
}

func (f *fakeHandlerData) AllActiveTasks(userID int) ([]models.Task, error) {
	if f.TasksErr != nil {
		return nil, f.TasksErr
	}
	_ = userID
	return f.Tasks, nil
}

func (f *fakeHandlerData) GetTask(taskID int) (models.Task, error) {
	if f.GetTaskErr != nil {
		return models.Task{}, f.GetTaskErr
	}
	if f.TasksByID == nil {
		return models.Task{}, sql.ErrNoRows
	}
	t, ok := f.TasksByID[taskID]
	if !ok {
		return models.Task{}, sql.ErrNoRows
	}
	return t, nil
}

func (f *fakeHandlerData) SaveTask(t models.Task) (int64, error) {
	if f.SaveTaskErr != nil {
		return 0, f.SaveTaskErr
	}
	if f.SaveTaskID != 0 {
		return f.SaveTaskID, nil
	}
	return 42, nil
}

func (f *fakeHandlerData) UpdateTask(t models.Task) error {
	if f.UpdateErr != nil {
		return f.UpdateErr
	}
	f.LastUpdateTask = t
	f.UpdateCount++
	if f.TasksByID != nil {
		f.TasksByID[t.ID] = t
	}
	return nil
}

func (f *fakeHandlerData) AllProjectsForUser(userID int) ([]models.Project, error) {
	if f.ProjectsErr != nil {
		return nil, f.ProjectsErr
	}
	_ = userID
	return f.Projects, nil
}

func (f *fakeHandlerData) GetProject(id int) (models.Project, error) {
	if f.ProjectsByID == nil {
		return models.Project{}, sql.ErrNoRows
	}
	p, ok := f.ProjectsByID[id]
	if !ok {
		return models.Project{}, sql.ErrNoRows
	}
	return p, nil
}

func (f *fakeHandlerData) SaveProject(models.Project) (int64, error) {
	return 1, nil
}

func (f *fakeHandlerData) UpdateProject(models.Project) error {
	return nil
}

func (f *fakeHandlerData) DeleteProjectForUser(int, int) error {
	return nil
}

func (f *fakeHandlerData) GetArea(id int) (models.Area, error) {
	if f.AreasByID == nil {
		return models.Area{}, sql.ErrNoRows
	}
	a, ok := f.AreasByID[id]
	if !ok {
		return models.Area{}, sql.ErrNoRows
	}
	return a, nil
}

func (f *fakeHandlerData) SaveArea(models.Area) (int64, error) {
	return 1, nil
}

func (f *fakeHandlerData) UpdateArea(models.Area) error {
	return nil
}

func (f *fakeHandlerData) DeleteAreaForUser(int, int) error {
	return nil
}

func (f *fakeHandlerData) AllContextsForUser(userID int) ([]models.Context, error) {
	_ = userID
	if f.Contexts != nil {
		return f.Contexts, nil
	}
	return nil, nil
}

func (f *fakeHandlerData) GetContext(id int) (models.Context, error) {
	if f.ContextsByID == nil {
		return models.Context{}, sql.ErrNoRows
	}
	c, ok := f.ContextsByID[id]
	if !ok {
		return models.Context{}, sql.ErrNoRows
	}
	return c, nil
}

func (f *fakeHandlerData) SaveContext(models.Context) (int64, error) {
	return 1, nil
}

func (f *fakeHandlerData) UpdateContext(models.Context) error {
	return nil
}

func (f *fakeHandlerData) DeleteContextForUser(int, int) error {
	return nil
}

func testApp(data HandlerData, auth Authenticator) *App {
	return &App{
		Data:      data,
		Auth:      auth,
		Templates: loadTemplates(),
	}
}

func requestWithUserID(r *http.Request, uid int) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userIDContextKey, uid))
}

func TestTasksListHandler_RedirectAndList(t *testing.T) {
	data := &fakeHandlerData{
		User: models.User{ID: 1, Name: "Test User", Inbox: "note"},
		Projects: []models.Project{{ID: 10, Name: "P1", UserID: 1}},
		Areas:    []models.Area{{ID: 20, Name: "A1", UserID: 1}},
		Tasks: []models.Task{
			{ID: 3, Name: "Alpha task", Status: models.StatusActive, UserID: 1, Priority: models.PriorityLow},
		},
	}
	auth := &fakeAuth{UID: 1, OK: true}
	app := testApp(data, auth)

	t.Run("redirect /tasks to /tasks/", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		app.requireAuth(app.tasksListHandler)(w, r)
		if w.Code != http.StatusSeeOther {
			t.Fatalf("status %d, want %d", w.Code, http.StatusSeeOther)
		}
		loc := w.Header().Get("Location")
		if loc != "/tasks/" {
			t.Fatalf("Location %q, want /tasks/", loc)
		}
	})

	t.Run("404 for /tasks/extra", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/extra", nil)
		app.requireAuth(app.tasksListHandler)(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status %d, want 404", w.Code)
		}
	})

	t.Run("200 lists task name", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/tasks/", nil)
		app.requireAuth(app.tasksListHandler)(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d, want 200", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Alpha task") {
			t.Fatalf("body should contain task name; got %q", body)
		}
		if !strings.Contains(body, "Active tasks") {
			t.Fatalf("body should contain page title")
		}
	})
}

func TestTasksListHandler_AuthRedirect(t *testing.T) {
	data := &fakeHandlerData{}
	app := testApp(data, &fakeAuth{OK: false})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/tasks/", nil)
	app.requireAuth(app.tasksListHandler)(w, r)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Fatalf("expected redirect to /login, got %d %q", w.Code, w.Header().Get("Location"))
	}
}

func TestAPITaskHandler_UpdateAndDone(t *testing.T) {
	task := models.Task{
		ID:          7,
		Name:        "Old",
		Status:      models.StatusActive,
		Priority:    models.PriorityLow,
		DateCreated: time.Now(),
		UserID:      5,
		ProjectID:   1,
		AreaID:      2,
	}
	data := &fakeHandlerData{
		TasksByID: map[int]models.Task{7: task},
	}
	auth := &fakeAuth{UID: 5, OK: true}
	app := testApp(data, auth)
	h := app.requireAuth(app.apiTaskHandler)

	t.Run("update success", func(t *testing.T) {
		body := `{"name":"New title","status":"active","priority":1,"project_id":3,"area_id":4}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/task/7/update", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		h(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d body %s", w.Code, w.Body.String())
		}
		var resp apiJSONResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if !resp.OK {
			t.Fatalf("expected ok: %+v", resp)
		}
		if data.LastUpdateTask.Name != "New title" {
			t.Fatalf("update name: got %q", data.LastUpdateTask.Name)
		}
		if data.LastUpdateTask.Priority != models.PriorityMed {
			t.Fatalf("priority: got %v", data.LastUpdateTask.Priority)
		}
		if data.LastUpdateTask.ProjectID != 3 || data.LastUpdateTask.AreaID != 4 {
			t.Fatalf("project/area: %+v", data.LastUpdateTask)
		}
	})

	t.Run("done success", func(t *testing.T) {
		data.UpdateCount = 0
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/task/7/done", nil)
		h(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d", w.Code)
		}
		if data.LastUpdateTask.Status != models.StatusDone {
			t.Fatalf("status %q", data.LastUpdateTask.Status)
		}
	})

	t.Run("wrong user", func(t *testing.T) {
		other := task
		other.UserID = 99
		data2 := &fakeHandlerData{TasksByID: map[int]models.Task{8: other}}
		app2 := testApp(data2, &fakeAuth{UID: 5, OK: true})
		h2 := app2.requireAuth(app2.apiTaskHandler)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/task/8/update", strings.NewReader(`{"name":"x","status":"active","priority":0}`))
		r.Header.Set("Content-Type", "application/json")
		h2(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status %d", w.Code)
		}
	})

	t.Run("missing task", func(t *testing.T) {
		data3 := &fakeHandlerData{TasksByID: map[int]models.Task{}}
		app3 := testApp(data3, auth)
		h3 := app3.requireAuth(app3.apiTaskHandler)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/task/999/update", strings.NewReader(`{"name":"x","status":"active","priority":0}`))
		r.Header.Set("Content-Type", "application/json")
		h3(w, r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/task/7/update", strings.NewReader(`not json`))
		r.Header.Set("Content-Type", "application/json")
		h(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status %d", w.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/task/7/update", nil)
		h(w, r)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status %d", w.Code)
		}
	})
}

func TestAPITaskHandler_RequiresAuth(t *testing.T) {
	data := &fakeHandlerData{TasksByID: map[int]models.Task{1: {ID: 1, Name: "x", UserID: 1}}}
	app := testApp(data, &fakeAuth{OK: false})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/task/1/done", nil)
	app.requireAuth(app.apiTaskHandler)(w, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", w.Code)
	}
}

// Ensures handlers that skip requireAuth still work if user id is injected (e.g. tests calling handler directly).
func TestUserIDFromRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = requestWithUserID(r, 123)
	if userIDFromRequest(r) != 123 {
		t.Fatalf("got %d", userIDFromRequest(r))
	}
}

var _ HandlerData = (*fakeHandlerData)(nil)
var _ Authenticator = (*fakeAuth)(nil)
