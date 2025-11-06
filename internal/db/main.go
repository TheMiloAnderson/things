package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

type Table struct {
	db     *sql.DB
	Tx     *sql.Tx
	DBName string
}

func (t *Table) Connect() {
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = t.DBName
	cfg.ParseTime = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}
	fmt.Println("Connected to DB:", t.DBName)
	t.db = db
}

func (t *Table) BeginTx() error {
	if t.db == nil {
		return fmt.Errorf("DB is not connected")
	}
	tx, err := t.db.Begin()
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
	return t.db.Exec(query, args...)
}

func (t *Table) QueryRow(query string, args ...interface{}) *sql.Row {
	if t.Tx != nil {
		return t.Tx.QueryRow(query, args...)
	}
	return t.db.QueryRow(query, args...)
}
