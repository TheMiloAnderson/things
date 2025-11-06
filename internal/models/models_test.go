package models

import (
	"database/sql"
	"testing"
	"time"
)

var dbName = "tasks_test"

func TestUserGetByID(t *testing.T) {
	u := User{}
	u.DBName = dbName
	u.Connect()
	if err := u.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	defer u.Rollback()

	err := u.GetById(1)
	if err != nil || u.Name != "Milo Anderson" {
		t.Errorf(`User GetByID Error: %v, Name: %v`, err, u.Name)
	}
}

func TestTaskSave(t *testing.T) {
	task := Task{}
	task.DBName = dbName
	task.Connect()

	if err := task.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	defer task.Rollback()

	task.Name = "Call Dave"
	task.Status = "active"
	task.Priority = 1
	task.DateCreated = time.Now()
	task.ProjectID = 1
	task.AreaID = 1
	task.UserID = 1

	taskID, err := task.Save()
	if err != nil {
		t.Fatalf("Task Save Error: %v", err)
	}

	newTask := Task{Table: task.Table}
	row := task.QueryRow("SELECT * FROM tasks WHERE id = ?", taskID)
	err = row.Scan(
		&newTask.ID,
		&newTask.Name,
		&newTask.Status,
		&newTask.Priority,
		&newTask.DateCreated,
		&newTask.ProjectID,
		&newTask.AreaID,
		&newTask.UserID,
	)
	if err != nil || newTask.Name != "Call Dave" {
		t.Fatalf("Task Save Error: %v; New Task: %+v", err, newTask)
	}
}

func TestTaskUpdate(t *testing.T) {
	task := Task{}
	task.DBName = dbName
	task.Connect()
	if err := task.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	defer task.Rollback()

	err := task.GetById(1)
	if err != nil || task.Name != "Call Aviator" {
		t.Errorf(`TestTaskUpdate GetByID Error: %v`, err)
	}

	task.Name = "Call Aviator Audio"
	err = task.Update()
	if err != nil {
		t.Errorf(`TestTaskUpdate Error: %v`, err)
	}

	task.GetById(1)
	if err != nil || task.Name != "Call Aviator Audio" {
		t.Errorf(`TestTaskUpdate GetByID Error: %v, Task: %v`, err, task)
	}
}

func TestTaskDelete(t *testing.T) {
	task := Task{}
	task.DBName = dbName
	task.Connect()
	if err := task.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	defer task.Rollback()

	task.ID = 1
	if err := task.Delete(); err != nil {
		t.Errorf(`Task Delete Error: %v`, err)
	}

	err := task.GetById(1)
	if err != sql.ErrNoRows {
		t.Errorf(`Task Delete - expected ErrNoRows but got: %v`, err)
	}
}
