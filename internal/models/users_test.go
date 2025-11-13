package models

import "testing"

func beginUserTransaction(t *testing.T) User {
	u := User{}
	u.DB = testDB
	if err := u.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	return u
}

func TestUserGetByID(t *testing.T) {
	u := beginUserTransaction(t)
	defer u.Rollback()

	err := u.GetById(1)
	if err != nil || u.Name != "Milo Anderson" {
		t.Errorf(`User GetByID Error: %v, Name: %v`, err, u.Name)
	}
}

func TestUserUpdate(t *testing.T) {
	u := beginUserTransaction(t)
	defer u.Rollback()

	err := u.GetById(1)
	if err != nil || u.Inbox != "inbox text has a 1:1 relationship with user!" {
		t.Errorf(`TestUserUpdate GetByID Error: %v`, err)
	}

	u.Inbox = "wut"
	err = u.Update()
	if err != nil {
		t.Errorf(`TestUserUpdate Error: %v`, err)
	}

	u.GetById(1)
	if err != nil || u.Inbox != "wut" {
		t.Errorf(`TestUserUpdate GetByID Error: %v, User: %v`, err, u)
	}
}
