package models

import (
	"database/sql"
	"os"
	"testing"
	"things/internal/db"
)

var testDB *sql.DB
var dbName = "tasks_test"

func TestMain(m *testing.M) {
	table := db.Table{}
	table.Connect(dbName)
	testDB = table.DB

	// Run all tests
	code := m.Run()

	testDB.Close()
	os.Exit(code)
}
