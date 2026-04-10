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
	StatusPending  Status = "pending"
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

// AllActiveForUser returns active and pending tasks for the given user_id (not done/canceled), newest first.
func (t *Task) AllActiveForUser(userID int) ([]Task, error) {
	rows, err := t.Query(
		`SELECT id, name, status, priority, date_created, project_id, area_id, user_id FROM tasks
		WHERE user_id = ? AND status IN (?, ?) ORDER BY date_created DESC`,
		userID, StatusActive, StatusPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		task.Connection = t.Connection
		var projectID sql.NullInt64
		var areaID sql.NullInt64
		err := rows.Scan(&task.ID, &task.Name, &task.Status, &task.Priority, &task.DateCreated, &projectID, &areaID, &task.UserID)
		if err != nil {
			return nil, err
		}
		if projectID.Valid {
			task.ProjectID = int(projectID.Int64)
		}
		if areaID.Valid {
			task.AreaID = int(areaID.Int64)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// AllTasksForProject returns every task for the user on the given project (any status), newest first.
func (t *Task) AllTasksForProject(userID, projectID int) ([]Task, error) {
	rows, err := t.Query(
		`SELECT id, name, status, priority, date_created, project_id, area_id, user_id FROM tasks
		WHERE user_id = ? AND project_id = ? ORDER BY date_created DESC`,
		userID, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		task.Connection = t.Connection
		var projectIDCol sql.NullInt64
		var areaIDCol sql.NullInt64
		err := rows.Scan(&task.ID, &task.Name, &task.Status, &task.Priority, &task.DateCreated, &projectIDCol, &areaIDCol, &task.UserID)
		if err != nil {
			return nil, err
		}
		if projectIDCol.Valid {
			task.ProjectID = int(projectIDCol.Int64)
		}
		if areaIDCol.Valid {
			task.AreaID = int(areaIDCol.Int64)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// CancelActiveTasksForProject sets status to canceled for active or pending tasks on this project.
// Tasks already done or canceled are left unchanged.
func (t *Task) CancelActiveTasksForProject(projectID, userID int) error {
	_, err := t.Exec(
		`UPDATE tasks SET status = ? WHERE project_id = ? AND user_id = ? AND status IN (?, ?)`,
		StatusCanceled, projectID, userID, StatusActive, StatusPending,
	)
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
