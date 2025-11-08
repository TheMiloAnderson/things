package models

import (
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
	row := t.QueryRow("SELECT * FROM tasks WHERE id = ?", id)
	err := row.Scan(&t.ID, &t.Name, &t.Status, &t.Priority, &t.DateCreated, &t.ProjectID, &t.AreaID, &t.UserID)
	return err
}

func (t *Task) Save() (int64, error) {
	// TODO figure out the ContextIDs logic
	result, err := t.Exec(
		`INSERT INTO tasks (name, status, priority, date_created, project_id, area_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, t.Name, t.Status, t.Priority, t.DateCreated, t.ProjectID, t.AreaID, t.UserID,
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
	_, err := t.Exec(
		`UPDATE tasks SET name = ?, status = ?, priority = ?, date_created = ?, project_id = ?, area_id = ?, user_id = ? 
		WHERE id = ?`, t.Name, t.Status, t.Priority, t.DateCreated, t.ProjectID, t.AreaID, t.UserID, t.ID,
	)
	return err
}

func (t *Task) Delete() error {
	_, err := t.Exec(`DELETE FROM tasks WHERE id = ?`, t.ID)
	return err
}
