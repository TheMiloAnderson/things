package models

import (
	"database/sql"
	"testing"
)

func beginAreaTransaction(t *testing.T) Area {
	a := Area{}
	a.DB = testDB
	if err := a.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	return a
}

func TestAreaGetByID(t *testing.T) {
	a := beginAreaTransaction(t)
	defer a.Rollback()

	err := a.GetById(1)
	if err != nil || a.Name != "Music" {
		t.Errorf(`TestAreaGetByID Error: %v`, err)
	}
}

func TestAreaSave(t *testing.T) {
	a := beginAreaTransaction(t)
	defer a.Rollback()

	a.Name = "Career"
	a.UserID = 1
	areaID, err := a.Save()
	if err != nil {
		t.Fatalf("TestAreaSave Error: %v", err)
	}

	newArea := Area{}
	row := a.QueryRow("SELECT * FROM areas WHERE id = ?", areaID)
	err = row.Scan(&newArea.ID, &newArea.Name, &newArea.UserID)
	if err != nil || newArea.Name != "Career" {
		t.Fatalf("TestAreaSave: %v; New Area: %+v", err, newArea)
	}
}

func TestAreaUpdate(t *testing.T) {
	a := beginAreaTransaction(t)
	defer a.Rollback()

	err := a.GetById(1)
	if err != nil || a.Name != "Music" {
		t.Errorf(`TestAreaUpdate GetByID Error: %v`, err)
	}

	a.Name = "Creativity"
	err = a.Update()
	if err != nil {
		t.Errorf(`TestAreaUpdate Error: %v`, err)
	}

	a.GetById(1)
	if err != nil || a.Name != "Creativity" {
		t.Errorf(`TestAreaUpdate GetByID Error: %v, Area: %v`, err, a)
	}
}

func TestAreaDelete(t *testing.T) {
	a := beginAreaTransaction(t)
	defer a.Rollback()

	a.ID = 1
	if err := a.Delete(); err != nil {
		t.Errorf(`Area Delete Error: %v`, err)
	}

	err := a.GetById(1)
	if err != sql.ErrNoRows {
		t.Errorf(`Area Delete - expected ErrNoRows but got: %v`, err)
	}
}
