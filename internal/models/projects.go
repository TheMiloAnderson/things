package models

import (
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
	err := row.Scan(&p.ID, &p.Name, &p.Status, &p.Notes, &p.AreaID, &p.UserID)
	return err
}

func (p *Project) Save() (int64, error) {
	result, err := p.Exec(
		`INSERT INTO projects (name, status, notes, area_id, user_id)
		VALUES (?, ?, ?, ?, ?)`, p.Name, p.Status, p.Notes, p.AreaID, p.UserID,
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
	_, err := p.Exec(
		`UPDATE projects SET name = ?, status = ?, notes = ?, area_id = ?, user_id = ? WHERE id = ?`,
		p.Name, p.Status, p.Notes, p.AreaID, p.UserID, p.ID,
	)
	return err
}

func (p *Project) Delete() error {
	_, err := p.Exec(`DELETE FROM projects WHERE id = ?`, p.ID)
	return err
}
