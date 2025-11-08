package models

import (
	"database/sql"
	"testing"
)

// Helper to start a transaction for Project tests
func beginProjectTransaction(t *testing.T) Project {
	p := Project{}
	p.DB = testDB
	if err := p.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	return p
}

func TestProjectGetByID(t *testing.T) {
	p := beginProjectTransaction(t)
	defer p.Rollback()

	err := p.GetById(1)
	if err != nil || p.Name != "Get amp repaired" {
		t.Errorf(`TestProjectGetByID Error: %v, Project: %+v`, err, p)
	}
}

func TestProjectSave(t *testing.T) {
	p := beginProjectTransaction(t)
	defer p.Rollback()

	p.Name = "New Website"
	p.Status = "active"
	p.Notes = "Build a new portfolio site"
	p.AreaID = 1
	p.UserID = 1

	projectID, err := p.Save()
	if err != nil {
		t.Fatalf("TestProjectSave Error: %v", err)
	}

	newProject := Project{}
	row := p.QueryRow("SELECT id, name, status, notes, area_id, user_id FROM projects WHERE id = ?", projectID)
	err = row.Scan(&newProject.ID, &newProject.Name, &newProject.Status, &newProject.Notes, &newProject.AreaID, &newProject.UserID)
	if err != nil || newProject.Name != "New Website" {
		t.Fatalf("TestProjectSave: %v; New Project: %+v", err, newProject)
	}
}

func TestProjectUpdate(t *testing.T) {
	p := beginProjectTransaction(t)
	defer p.Rollback()

	err := p.GetById(1)
	if err != nil {
		t.Fatalf("TestProjectUpdate GetByID Error: %v", err)
	}

	p.Name = "Fix amp completely"
	p.Notes = "Call Aviator Audio for full diagnostic"
	err = p.Update()
	if err != nil {
		t.Errorf(`TestProjectUpdate Error: %v`, err)
	}

	err = p.GetById(1)
	if err != nil || p.Name != "Fix amp completely" {
		t.Errorf(`TestProjectUpdate GetByID validation failed: %v, Project: %+v`, err, p)
	}
}

func TestProjectDelete(t *testing.T) {
	p := beginProjectTransaction(t)
	defer p.Rollback()

	p.ID = 1
	if err := p.Delete(); err != nil {
		t.Errorf(`Project Delete Error: %v`, err)
	}

	err := p.GetById(1)
	if err != sql.ErrNoRows {
		t.Errorf(`Project Delete - expected ErrNoRows but got: %v`, err)
	}
}
