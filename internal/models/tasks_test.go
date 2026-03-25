package models

import (
	"database/sql"
	"testing"
	"time"
)

func beginTaskTransaction(t *testing.T) Task {
	task := Task{}
	task.DB = testDB
	if err := task.BeginTx(); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
	return task
}

func TestTaskGetByID(t *testing.T) {
	task := beginTaskTransaction(t)
	defer task.Rollback()

	err := task.GetById(1)
	if err != nil || task.Name != "Call Aviator" {
		t.Errorf(`TestTaskGetByID Error: %v`, err)
	}
}

func TestTaskSave(t *testing.T) {
	task := beginTaskTransaction(t)
	defer task.Rollback()

	task.Name = "Call Dave"
	task.Status = StatusActive
	task.Priority = PriorityMed
	task.DateCreated = time.Now()
	task.ProjectID = 1
	task.AreaID = 1
	task.UserID = 1

	taskID, err := task.Save()
	if err != nil {
		t.Fatalf("Task Save Error: %v", err)
	}

	newTask := Task{}
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
	task := beginTaskTransaction(t)
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
	task := beginTaskTransaction(t)
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

func TestTaskAllActiveForUser(t *testing.T) {
	task := beginTaskTransaction(t)
	defer task.Rollback()

	// test_data: user 1 has one active task ("Call Aviator"); add another active and one done
	t2 := Task{Connection: task.Connection}
	t2.Name = "Second active"
	t2.Status = StatusActive
	t2.Priority = PriorityLow
	t2.DateCreated = time.Now()
	t2.ProjectID = 1
	t2.AreaID = 1
	t2.UserID = 1
	if _, err := t2.Save(); err != nil {
		t.Fatalf("Save active task: %v", err)
	}

	t3 := Task{Connection: task.Connection}
	t3.Name = "Done task"
	t3.Status = StatusDone
	t3.Priority = PriorityLow
	t3.DateCreated = time.Now()
	t3.ProjectID = 1
	t3.AreaID = 1
	t3.UserID = 1
	if _, err := t3.Save(); err != nil {
		t.Fatalf("Save done task: %v", err)
	}

	list, err := task.AllActiveForUser(1)
	if err != nil {
		t.Fatalf("AllActiveForUser: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("Expected 2 active tasks for user 1, got %d", len(list))
	}
	for _, tk := range list {
		if tk.Status != StatusActive {
			t.Errorf("Non-active task in list: %+v", tk)
		}
		if tk.UserID != 1 {
			t.Errorf("Wrong user: %+v", tk)
		}
	}
}
