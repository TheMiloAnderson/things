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

func TestAreaAllForUser(t *testing.T) {
	a := beginAreaTransaction(t)
	defer a.Rollback()

	// Insert multiple areas for user 1; test_data has one user with id=1 and one area "Music"
	areaNames := []string{"Writing", "Cooking", "Fitness"}
	for _, name := range areaNames {
		temp := Area{Connection: a.Connection, Name: name, UserID: 1}
		_, err := temp.Save()
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	areas, err := a.AllForUser(1)
	if err != nil {
		t.Fatalf("AllForUser error: %v", err)
	}
	if len(areas) != 4 {
		t.Fatalf("Expected 4 areas for user 1, got %d", len(areas))
	}
	for _, ar := range areas {
		if ar.UserID != 1 {
			t.Errorf("Wrong user area returned: %+v", ar)
		}
	}
	// Check our three names are present
	seen := make(map[string]bool)
	for _, ar := range areas {
		seen[ar.Name] = true
	}
	for _, name := range areaNames {
		if !seen[name] {
			t.Errorf("Expected area %q in results", name)
		}
	}
}
