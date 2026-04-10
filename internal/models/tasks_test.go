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

	// test_data: user 1 has one active task ("Call Aviator"); add another active, one pending, and one done
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

	tPending := Task{Connection: task.Connection}
	tPending.Name = "Planned later"
	tPending.Status = StatusPending
	tPending.Priority = PriorityLow
	tPending.DateCreated = time.Now()
	tPending.ProjectID = 1
	tPending.AreaID = 1
	tPending.UserID = 1
	if _, err := tPending.Save(); err != nil {
		t.Fatalf("Save pending task: %v", err)
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
	if len(list) != 3 {
		t.Fatalf("Expected 3 open tasks for user 1, got %d", len(list))
	}
	for _, tk := range list {
		if tk.Status != StatusActive && tk.Status != StatusPending {
			t.Errorf("Task in open list should be active or pending: %+v", tk)
		}
		if tk.UserID != 1 {
			t.Errorf("Wrong user: %+v", tk)
		}
	}
}

func TestTaskCancelActiveTasksForProject(t *testing.T) {
	task := beginTaskTransaction(t)
	defer task.Rollback()

	active := Task{Connection: task.Connection}
	active.Name = "Cascade me"
	active.Status = StatusActive
	active.Priority = PriorityLow
	active.DateCreated = time.Now()
	active.ProjectID = 1
	active.AreaID = 1
	active.UserID = 1
	activeID, err := active.Save()
	if err != nil {
		t.Fatalf("Save active: %v", err)
	}

	done := Task{Connection: task.Connection}
	done.Name = "Leave done"
	done.Status = StatusDone
	done.Priority = PriorityLow
	done.DateCreated = time.Now()
	done.ProjectID = 1
	done.AreaID = 1
	done.UserID = 1
	doneID, err := done.Save()
	if err != nil {
		t.Fatalf("Save done: %v", err)
	}

	pending := Task{Connection: task.Connection}
	pending.Name = "Cascade pending"
	pending.Status = StatusPending
	pending.Priority = PriorityLow
	pending.DateCreated = time.Now()
	pending.ProjectID = 1
	pending.AreaID = 1
	pending.UserID = 1
	pendingID, err := pending.Save()
	if err != nil {
		t.Fatalf("Save pending: %v", err)
	}

	if err := task.CancelActiveTasksForProject(1, 1); err != nil {
		t.Fatalf("CancelActiveTasksForProject: %v", err)
	}

	if err := task.GetById(int(activeID)); err != nil {
		t.Fatalf("GetById active: %v", err)
	}
	if task.Status != StatusCanceled {
		t.Errorf("Active task should be canceled, got %q", task.Status)
	}

	verifyDone := Task{Connection: task.Connection}
	if err := verifyDone.GetById(int(doneID)); err != nil {
		t.Fatalf("GetById done: %v", err)
	}
	if verifyDone.Status != StatusDone {
		t.Errorf("Done task should stay done, got %q", verifyDone.Status)
	}

	verifyPending := Task{Connection: task.Connection}
	if err := verifyPending.GetById(int(pendingID)); err != nil {
		t.Fatalf("GetById pending: %v", err)
	}
	if verifyPending.Status != StatusCanceled {
		t.Errorf("Pending task should be canceled, got %q", verifyPending.Status)
	}
}

func TestTaskAllTasksForProject(t *testing.T) {
	task := beginTaskTransaction(t)
	defer task.Rollback()

	before, err := task.AllTasksForProject(1, 1)
	if err != nil {
		t.Fatalf("AllTasksForProject: %v", err)
	}
	n0 := len(before)

	tA := Task{Connection: task.Connection}
	tA.Name = "AllProj active"
	tA.Status = StatusActive
	tA.Priority = PriorityLow
	tA.DateCreated = time.Now()
	tA.ProjectID = 1
	tA.AreaID = 1
	tA.UserID = 1
	if _, err := tA.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	tD := Task{Connection: task.Connection}
	tD.Name = "AllProj done"
	tD.Status = StatusDone
	tD.Priority = PriorityLow
	tD.DateCreated = time.Now()
	tD.ProjectID = 1
	tD.AreaID = 1
	tD.UserID = 1
	if _, err := tD.Save(); err != nil {
		t.Fatalf("Save done: %v", err)
	}

	after, err := task.AllTasksForProject(1, 1)
	if err != nil {
		t.Fatalf("AllTasksForProject: %v", err)
	}
	if len(after) != n0+2 {
		t.Fatalf("expected %d tasks, got %d", n0+2, len(after))
	}
	for _, tk := range after {
		if tk.UserID != 1 || tk.ProjectID != 1 {
			t.Errorf("Wrong scope: %+v", tk)
		}
	}
}
