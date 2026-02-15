package models

import (
	"database/sql"
	"things/internal/db"
	"time"
)

type Task struct {
	db.Connection
	ID          int
	Name        string
	Status      Status
	Priority    Priority
	DateCreated time.Time
	ProjectID   int
	AreaID      int
	UserID      int
	ContextIDs  []int
}

type Status string

const (
	StatusActive   Status = "active"
	StatusDone     Status = "done"
	StatusCanceled Status = "canceled"
)

type Priority int

const (
	PriorityLow  Priority = 0
	PriorityMed  Priority = 1
	PriorityHigh Priority = 2
)

func (t *Task) GetById(id int) error {
	row := t.QueryRow("SELECT * FROM tasks WHERE id = ?", id)
	var projectID sql.NullInt64
	var areaID sql.NullInt64
	err := row.Scan(&t.ID, &t.Name, &t.Status, &t.Priority, &t.DateCreated, &projectID, &areaID, &t.UserID)
	if err != nil {
		return err
	}
	if projectID.Valid {
		t.ProjectID = int(projectID.Int64)
	} else {
		t.ProjectID = 0
	}
	if areaID.Valid {
		t.AreaID = int(areaID.Int64)
	} else {
		t.AreaID = 0
	}
	return err
}

func (t *Task) Save() (int64, error) {
	// TODO figure out the ContextIDs logic
	var projectID any = t.ProjectID
	if t.ProjectID == 0 {
		projectID = nil
	}
	var areaID any = t.AreaID
	if t.AreaID == 0 {
		areaID = nil
	}
	result, err := t.Exec(
		`INSERT INTO tasks (name, status, priority, date_created, project_id, area_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, t.Name, t.Status, t.Priority, t.DateCreated, projectID, areaID, t.UserID,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (t *Task) Update() error {
	var projectID any = t.ProjectID
	if t.ProjectID == 0 {
		projectID = nil
	}
	var areaID any = t.AreaID
	if t.AreaID == 0 {
		areaID = nil
	}
	_, err := t.Exec(
		`UPDATE tasks SET name = ?, status = ?, priority = ?, date_created = ?, project_id = ?, area_id = ?, user_id = ? 
		WHERE id = ?`, t.Name, t.Status, t.Priority, t.DateCreated, projectID, areaID, t.UserID, t.ID,
	)
	return err
}

func (t *Task) Delete() error {
	_, err := t.Exec(`DELETE FROM tasks WHERE id = ?`, t.ID)
	return err
}
