package models

import (
	"database/sql"
	"fmt"
	"testing"
)

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
	p.Status = StatusActive
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

func TestProjectAllActiveForUser(t *testing.T) {
	p := beginProjectTransaction(t)
	defer p.Rollback()

	// Insert multiple projects for user 1 (active/done/canceled); test_data has one user with id=1
	for i, status := range []Status{"active", "done", "active", "canceled"} {
		temp := Project{Connection: p.Connection, Name: fmt.Sprintf("Proj%d", i+1), Status: status, Notes: "", AreaID: 1, UserID: 1}
		_, err := temp.Save()
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}
	projects, err := p.AllActiveForUser(1)
	if err != nil {
		t.Fatalf("AllActiveForUser error: %v", err)
	}
	// test_data has 1 active project ("Get amp repaired"), we insert 2 more active (Proj1, Proj3)
	if len(projects) != 3 {
		t.Fatalf("Expected 3 active projects, got %d", len(projects))
	}
	for _, proj := range projects {
		if proj.Status != StatusActive {
			t.Errorf("Inactive project returned: %+v", proj)
		}
		if proj.UserID != 1 {
			t.Errorf("Wrong user project returned: %+v", proj)
		}
	}
}

func TestProjectAllForUser(t *testing.T) {
	p := beginProjectTransaction(t)
	defer p.Rollback()

	// Insert a done project for user 1; test_data already has one active
	pj := Project{Connection: p.Connection}
	pj.Name = "Archived thing"
	pj.Status = StatusDone
	pj.Notes = ""
	pj.AreaID = 1
	pj.UserID = 1
	if _, err := pj.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	all, err := p.AllForUser(1)
	if err != nil {
		t.Fatalf("AllForUser: %v", err)
	}
	if len(all) < 2 {
		t.Fatalf("Expected at least 2 projects for user 1, got %d", len(all))
	}
	var seenDone bool
	for _, pr := range all {
		if pr.UserID != 1 {
			t.Errorf("wrong user: %+v", pr)
		}
		if pr.Status == StatusDone && pr.Name == "Archived thing" {
			seenDone = true
		}
	}
	if !seenDone {
		t.Fatal("done project not in AllForUser results")
	}
}
