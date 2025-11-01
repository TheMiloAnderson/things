package models

import (
	"testing"
)

var dbName = "tasks_test"

func TestGetByID(t *testing.T) {
	u := User{}
	u.DBName = dbName
	u.Connect()
	tx, err := u.DB.Begin()
	if err != nil {
		t.Errorf(`Transaction Error: %v,`, err)
	}
	defer tx.Rollback()

	err = u.GetById(1)
	if err != nil || u.Name != "Milo Anderson" {
		t.Errorf(`User GetByID Error: %v, Name: %v`, err, u.Name)
	}
}
