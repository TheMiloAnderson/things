package main

import (
	"database/sql"
	"things/internal/db"
	"things/internal/models"
)

// HandlerData abstracts persistence for HTTP handlers (enables tests with fakes).
type HandlerData interface {
	GetUser(userID int) (models.User, error)
	UpdateUserInbox(userID int, inbox string) error
	AllActiveProjects(userID int) ([]models.Project, error)
	AllAreas(userID int) ([]models.Area, error)
	AllActiveTasks(userID int) ([]models.Task, error)
	GetTask(taskID int) (models.Task, error)
	SaveTask(t models.Task) (int64, error)
	UpdateTask(t models.Task) error

	AllProjectsForUser(userID int) ([]models.Project, error)
	GetProject(projectID int) (models.Project, error)
	SaveProject(p models.Project) (int64, error)
	UpdateProject(p models.Project) error
	DeleteProjectForUser(projectID, userID int) error

	GetArea(areaID int) (models.Area, error)
	SaveArea(a models.Area) (int64, error)
	UpdateArea(a models.Area) error
	DeleteAreaForUser(areaID, userID int) error

	AllContextsForUser(userID int) ([]models.Context, error)
	GetContext(contextID int) (models.Context, error)
	SaveContext(c models.Context) (int64, error)
	UpdateContext(c models.Context) error
	DeleteContextForUser(contextID, userID int) error
}

// sqlHandlerData is the production implementation backed by MySQL models.
type sqlHandlerData struct {
	conn *db.Connection
}

func (s *sqlHandlerData) GetUser(userID int) (models.User, error) {
	u := models.User{Connection: *s.conn}
	err := u.GetById(userID)
	return u, err
}

func (s *sqlHandlerData) UpdateUserInbox(userID int, inbox string) error {
	u := models.User{Connection: *s.conn}
	if err := u.GetById(userID); err != nil {
		return err
	}
	u.Inbox = inbox
	return u.Update()
}

func (s *sqlHandlerData) AllActiveProjects(userID int) ([]models.Project, error) {
	p := models.Project{Connection: *s.conn}
	return p.AllActiveForUser(userID)
}

func (s *sqlHandlerData) AllAreas(userID int) ([]models.Area, error) {
	a := models.Area{Connection: *s.conn}
	return a.AllForUser(userID)
}

func (s *sqlHandlerData) AllActiveTasks(userID int) ([]models.Task, error) {
	t := models.Task{Connection: *s.conn}
	return t.AllActiveForUser(userID)
}

func (s *sqlHandlerData) GetTask(taskID int) (models.Task, error) {
	t := models.Task{Connection: *s.conn}
	err := t.GetById(taskID)
	return t, err
}

func (s *sqlHandlerData) SaveTask(t models.Task) (int64, error) {
	t.Connection = *s.conn
	return t.Save()
}

func (s *sqlHandlerData) UpdateTask(t models.Task) error {
	t.Connection = *s.conn
	return t.Update()
}

func (s *sqlHandlerData) AllProjectsForUser(userID int) ([]models.Project, error) {
	p := models.Project{Connection: *s.conn}
	return p.AllForUser(userID)
}

func (s *sqlHandlerData) GetProject(projectID int) (models.Project, error) {
	p := models.Project{Connection: *s.conn}
	err := p.GetById(projectID)
	return p, err
}

func (s *sqlHandlerData) SaveProject(p models.Project) (int64, error) {
	p.Connection = *s.conn
	return p.Save()
}

func (s *sqlHandlerData) UpdateProject(p models.Project) error {
	p.Connection = *s.conn
	return p.Update()
}

func (s *sqlHandlerData) DeleteProjectForUser(projectID, userID int) error {
	p := models.Project{Connection: *s.conn}
	if err := p.GetById(projectID); err != nil {
		return err
	}
	if p.UserID != userID {
		return sql.ErrNoRows
	}
	return p.Delete()
}

func (s *sqlHandlerData) GetArea(areaID int) (models.Area, error) {
	a := models.Area{Connection: *s.conn}
	err := a.GetById(areaID)
	return a, err
}

func (s *sqlHandlerData) SaveArea(a models.Area) (int64, error) {
	a.Connection = *s.conn
	return a.Save()
}

func (s *sqlHandlerData) UpdateArea(a models.Area) error {
	a.Connection = *s.conn
	return a.Update()
}

func (s *sqlHandlerData) DeleteAreaForUser(areaID, userID int) error {
	a := models.Area{Connection: *s.conn}
	if err := a.GetById(areaID); err != nil {
		return err
	}
	if a.UserID != userID {
		return sql.ErrNoRows
	}
	return a.Delete()
}

func (s *sqlHandlerData) AllContextsForUser(userID int) ([]models.Context, error) {
	c := models.Context{Connection: *s.conn}
	return c.AllForUser(userID)
}

func (s *sqlHandlerData) GetContext(contextID int) (models.Context, error) {
	c := models.Context{Connection: *s.conn}
	err := c.GetById(contextID)
	return c, err
}

func (s *sqlHandlerData) SaveContext(c models.Context) (int64, error) {
	c.Connection = *s.conn
	return c.Save()
}

func (s *sqlHandlerData) UpdateContext(c models.Context) error {
	c.Connection = *s.conn
	return c.Update()
}

func (s *sqlHandlerData) DeleteContextForUser(contextID, userID int) error {
	c := models.Context{Connection: *s.conn}
	if err := c.GetById(contextID); err != nil {
		return err
	}
	if c.UserID != userID {
		return sql.ErrNoRows
	}
	return c.Delete()
}
