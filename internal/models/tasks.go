package models

import (
	"fmt"
	"things/internal/db"
	"time"
)

type Task struct {
	db.Table
	ID          int
	Name        string
	Status      string
	Priority    int
	DateCreated time.Time
	ProjectID   int
	AreaID      int
	UserID      int
	ContextIDs  []int
}

func (t *Task) GetById(id int) error {
	row := t.DB.QueryRow("SELECT * FROM tasks WHERE id = ?", id)
	err := row.Scan(&t.ID, &t.Name, &t.Status, &t.Priority, &t.DateCreated, &t.ProjectID, &t.AreaID, &t.UserID)
	return err
}

func (t *Task) Save() (int64, error) {
	// TODO figure out the ContextIDs logic
	result, err := t.DB.Exec(
		`INSERT INTO tasks (name, status, priority, date_created, project_id, area_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, t.Name, t.Status, t.Priority, t.DateCreated, t.ProjectID, t.AreaID, t.UserID,
	)
	if err != nil {
		return 0, fmt.Errorf("Save Task: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Save Task: %v", err)
	}
	return id, nil
}
