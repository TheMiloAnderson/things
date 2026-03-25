package main

import (
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
