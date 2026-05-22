package models

import (
	"database/sql"
	"strings"
	"things/internal/db"
	"time"
)

func dedupePositiveInts(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

type Task struct {
	db.Connection
	ID          int
	Name        string
	Status      Status
	Priority    Priority
	Notes       string
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
	var notes sql.NullString
	err := row.Scan(&t.ID, &t.Name, &t.Status, &t.Priority, &notes, &t.DateCreated, &projectID, &areaID, &t.UserID)
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
	if notes.Valid {
		t.Notes = notes.String
	}
	return err
}

// AllActiveForUser returns active and pending tasks for the given user_id (not done/canceled), newest first.
func (t *Task) AllActiveForUser(userID int) ([]Task, error) {
	rows, err := t.Query(
		`SELECT id, name, status, priority, notes, date_created, project_id, area_id, user_id FROM tasks
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
		var notes sql.NullString
		err := rows.Scan(&task.ID, &task.Name, &task.Status, &task.Priority, &notes, &task.DateCreated, &projectID, &areaID, &task.UserID)
		if err != nil {
			return nil, err
		}
		if projectID.Valid {
			task.ProjectID = int(projectID.Int64)
		}
		if areaID.Valid {
			task.AreaID = int(areaID.Int64)
		}
		if notes.Valid {
			task.Notes = notes.String
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// TaskFilter holds optional filter and sort criteria for listing tasks.
type TaskFilter struct {
	Status    Status // empty = active + pending (default)
	ProjectID int    // 0 = any
	AreaID    int    // 0 = any
	ContextID int    // 0 = any
	Sort      string // priority_desc | priority_asc | created_desc | created_asc
}

func scanTasksFromRows(rows *sql.Rows, conn db.Connection) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var task Task
		task.Connection = conn
		var projectID sql.NullInt64
		var areaID sql.NullInt64
		var notes sql.NullString
		err := rows.Scan(&task.ID, &task.Name, &task.Status, &task.Priority, &notes, &task.DateCreated, &projectID, &areaID, &task.UserID)
		if err != nil {
			return nil, err
		}
		if projectID.Valid {
			task.ProjectID = int(projectID.Int64)
		}
		if areaID.Valid {
			task.AreaID = int(areaID.Int64)
		}
		if notes.Valid {
			task.Notes = notes.String
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// FilteredTasks returns tasks matching all non-zero filter fields (AND logic).
func (t *Task) FilteredTasks(userID int, f TaskFilter) ([]Task, error) {
	var b strings.Builder
	b.WriteString(`SELECT DISTINCT t.id, t.name, t.status, t.priority, t.notes, t.date_created, t.project_id, t.area_id, t.user_id FROM tasks t`)

	args := make([]interface{}, 0, 8)
	if f.ContextID > 0 {
		b.WriteString(` INNER JOIN task_contexts tc ON tc.task_id = t.id AND tc.context_id = ?`)
		args = append(args, f.ContextID)
	}

	b.WriteString(` WHERE t.user_id = ?`)
	args = append(args, userID)

	if f.Status != "" {
		b.WriteString(` AND t.status = ?`)
		args = append(args, f.Status)
	} else {
		b.WriteString(` AND t.status IN (?, ?)`)
		args = append(args, StatusActive, StatusPending)
	}

	if f.ProjectID > 0 {
		b.WriteString(` AND t.project_id = ?`)
		args = append(args, f.ProjectID)
	}
	if f.AreaID > 0 {
		b.WriteString(` AND t.area_id = ?`)
		args = append(args, f.AreaID)
	}

	switch f.Sort {
	case "priority_desc":
		b.WriteString(` ORDER BY t.priority DESC, t.date_created DESC`)
	case "created_asc":
		b.WriteString(` ORDER BY t.date_created ASC`)
	default:
		b.WriteString(` ORDER BY t.date_created DESC`)
	}

	rows, err := t.Query(b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasksFromRows(rows, t.Connection)
}

// AllTasksForProject returns every task for the user on the given project (any status), newest first.
func (t *Task) AllTasksForProject(userID, projectID int) ([]Task, error) {
	rows, err := t.Query(
		`SELECT id, name, status, priority, notes, date_created, project_id, area_id, user_id FROM tasks
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
		var notes sql.NullString
		err := rows.Scan(&task.ID, &task.Name, &task.Status, &task.Priority, &notes, &task.DateCreated, &projectIDCol, &areaIDCol, &task.UserID)
		if err != nil {
			return nil, err
		}
		if projectIDCol.Valid {
			task.ProjectID = int(projectIDCol.Int64)
		}
		if areaIDCol.Valid {
			task.AreaID = int(areaIDCol.Int64)
		}
		if notes.Valid {
			task.Notes = notes.String
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

func (t *Task) insertTaskContexts(taskID int) error {
	for _, cid := range dedupePositiveInts(t.ContextIDs) {
		if _, err := t.Exec(
			`INSERT INTO task_contexts (task_id, context_id) VALUES (?, ?)`,
			taskID, cid,
		); err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) Save() (insertID int64, err error) {
	ownsTx := false
	if t.Tx == nil && t.DB != nil {
		if err := t.BeginTx(); err != nil {
			return 0, err
		}
		ownsTx = true
		defer func() {
			if err != nil && ownsTx {
				_ = t.Rollback()
			}
		}()
	}

	var projectID any = t.ProjectID
	if t.ProjectID == 0 {
		projectID = nil
	}
	var areaID any = t.AreaID
	if t.AreaID == 0 {
		areaID = nil
	}
	result, err := t.Exec(
		`INSERT INTO tasks (name, status, priority, notes, date_created, project_id, area_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, t.Name, t.Status, t.Priority, t.Notes, t.DateCreated, projectID, areaID, t.UserID,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err = t.insertTaskContexts(int(id)); err != nil {
		return 0, err
	}
	if ownsTx {
		if err = t.Commit(); err != nil {
			return 0, err
		}
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
		`UPDATE tasks SET name = ?, status = ?, priority = ?, notes = ?, date_created = ?, project_id = ?, area_id = ?, user_id = ? 
		WHERE id = ?`, t.Name, t.Status, t.Priority, t.Notes, t.DateCreated, projectID, areaID, t.UserID, t.ID,
	)
	return err
}

func (t *Task) Delete() error {
	_, err := t.Exec(`DELETE FROM tasks WHERE id = ?`, t.ID)
	return err
}
