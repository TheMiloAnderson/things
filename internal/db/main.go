package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Table struct {
	DB *sql.DB
	Tx *sql.Tx
}

func (t *Table) Connect(dbName string) {
	if t.DB != nil {
		// already connected (e.g., set by tests)
		return
	}
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = dbName
	cfg.ParseTime = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}
	fmt.Printf("✅ Connected to %v database\n", dbName)
	t.DB = db
}

func (t *Table) BeginTx() error {
	if t.DB == nil {
		return fmt.Errorf("DB is not connected")
	}
	tx, err := t.DB.Begin()
	if err != nil {
		return err
	}
	t.Tx = tx
	return nil
}

func (t *Table) Commit() error {
	if t.Tx == nil {
		return fmt.Errorf("no active transaction")
	}
	err := t.Tx.Commit()
	t.Tx = nil
	return err
}

func (t *Table) Rollback() error {
	if t.Tx == nil {
		return fmt.Errorf("no active transaction")
	}
	err := t.Tx.Rollback()
	t.Tx = nil
	return err
}

func (t *Table) Exec(query string, args ...interface{}) (sql.Result, error) {
	if t.Tx != nil {
		return t.Tx.Exec(query, args...)
	}
	return t.DB.Exec(query, args...)
}

func (t *Table) QueryRow(query string, args ...interface{}) *sql.Row {
	if t.Tx != nil {
		return t.Tx.QueryRow(query, args...)
	}
	return t.DB.QueryRow(query, args...)
}
