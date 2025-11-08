package models

import (
	"database/sql"
	"testing"
)

// Helper to start a transaction for Context tests
func beginContextTransaction(t *testing.T) Context {
	c := Context{}
	c.DB = testDB
	if err := c.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	return c
}

func TestContextGetByID(t *testing.T) {
	c := beginContextTransaction(t)
	defer c.Rollback()

	err := c.GetById(1)
	if err != nil || c.Name != "Business Hours" {
		t.Errorf(`TestContextGetByID Error: %v, Context: %+v`, err, c)
	}
}

func TestContextSave(t *testing.T) {
	c := beginContextTransaction(t)
	defer c.Rollback()

	c.Name = "Home Office"
	c.UserID = 1
	contextID, err := c.Save()
	if err != nil {
		t.Fatalf("TestContextSave Error: %v", err)
	}

	newContext := Context{}
	row := c.QueryRow("SELECT * FROM contexts WHERE id = ?", contextID)
	err = row.Scan(&newContext.ID, &newContext.Name, &newContext.UserID)
	if err != nil || newContext.Name != "Home Office" {
		t.Fatalf("TestContextSave: %v; New Context: %+v", err, newContext)
	}
}

func TestContextUpdate(t *testing.T) {
	c := beginContextTransaction(t)
	defer c.Rollback()

	err := c.GetById(1)
	if err != nil {
		t.Fatalf("TestContextUpdate GetByID Error: %v", err)
	}

	c.Name = "Snagglepuss"
	err = c.Update()
	if err != nil {
		t.Errorf("TestContextUpdate Error: %v", err)
	}

	err = c.GetById(1)
	if err != nil || c.Name != "Snagglepuss" {
		t.Errorf(`TestContextUpdate validation failed: %v, Context: %+v`, err, c)
	}
}

func TestContextDelete(t *testing.T) {
	c := beginContextTransaction(t)
	defer c.Rollback()

	c.ID = 1
	if err := c.Delete(); err != nil {
		t.Errorf(`Context Delete Error: %v`, err)
	}

	err := c.GetById(1)
	if err != sql.ErrNoRows {
		t.Errorf(`Context Delete - expected ErrNoRows but got: %v`, err)
	}
}
