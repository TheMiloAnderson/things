package models

import (
	"database/sql"
	"os"
	"testing"
	"things/internal/db"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	os.Chdir("../../")
	conn := db.Connection{}
	conn.Connect("tasks_test")
	testDB = conn.DB

	// Run all tests
	code := m.Run()

	testDB.Close()
	os.Exit(code)
}
