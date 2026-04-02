package models

import (
	"database/sql"
	"things/internal/db"
)

type Project struct {
	db.Connection
	ID     int
	Name   string
	Status Status
	Notes  string
	AreaID int
	UserID int
}

func (p *Project) GetById(id int) error {
	row := p.QueryRow("SELECT * FROM projects WHERE id = ?", id)
	var areaID sql.NullInt64
	err := row.Scan(&p.ID, &p.Name, &p.Status, &p.Notes, &areaID, &p.UserID)
	if err != nil {
		return err
	}
	if areaID.Valid {
		p.AreaID = int(areaID.Int64)
	} else {
		p.AreaID = 0
	}
	return nil
}

func (p *Project) Save() (int64, error) {
	var areaID any = p.AreaID
	if p.AreaID == 0 {
		areaID = nil
	}
	result, err := p.Exec(
		`INSERT INTO projects (name, status, notes, area_id, user_id)
		VALUES (?, ?, ?, ?, ?)`, p.Name, p.Status, p.Notes, areaID, p.UserID,
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

func (p *Project) Update() error {
	var areaID any = p.AreaID
	if p.AreaID == 0 {
		areaID = nil
	}
	_, err := p.Exec(
		`UPDATE projects SET name = ?, status = ?, notes = ?, area_id = ?, user_id = ? WHERE id = ?`,
		p.Name, p.Status, p.Notes, areaID, p.UserID, p.ID,
	)
	return err
}

func (p *Project) Delete() error {
	_, err := p.Exec(`DELETE FROM projects WHERE id = ?`, p.ID)
	return err
}

func (p *Project) AllActiveForUser(userID int) ([]Project, error) {
	rows, err := p.Query("SELECT id, name, status, notes, area_id, user_id FROM projects WHERE user_id = ? AND status = 'active' ORDER BY name ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []Project
	for rows.Next() {
		var prj Project
		err := rows.Scan(&prj.ID, &prj.Name, &prj.Status, &prj.Notes, &prj.AreaID, &prj.UserID)
		if err != nil {
			return nil, err
		}
		projects = append(projects, prj)
	}
	return projects, nil
}

// AllForUser returns every project for the user (any status), ordered by name.
func (p *Project) AllForUser(userID int) ([]Project, error) {
	rows, err := p.Query(
		`SELECT id, name, status, notes, area_id, user_id FROM projects WHERE user_id = ? ORDER BY name ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []Project
	for rows.Next() {
		var prj Project
		var areaID sql.NullInt64
		err := rows.Scan(&prj.ID, &prj.Name, &prj.Status, &prj.Notes, &areaID, &prj.UserID)
		if err != nil {
			return nil, err
		}
		if areaID.Valid {
			prj.AreaID = int(areaID.Int64)
		}
		projects = append(projects, prj)
	}
	return projects, rows.Err()
}
