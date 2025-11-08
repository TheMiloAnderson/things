package models

import "testing"

func TestUserGetByID(t *testing.T) {
	u := User{}
	u.DB = testDB
	if err := u.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	defer u.Rollback()

	err := u.GetById(1)
	if err != nil || u.Name != "Milo Anderson" {
		t.Errorf(`User GetByID Error: %v, Name: %v`, err, u.Name)
	}
}
