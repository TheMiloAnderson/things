package models

import "things/internal/db"

type Project struct {
	db.Table
	ID     int
	Name   string
	Status string
	Notes  string
	AreaID int
	UserID int
}

func (p *Project) GetById(id int) error {
	row := p.DB.QueryRow("SELECT * FROM projects WHERE id = ?", id)
	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Status,
		&p.Notes,
		&p.AreaID,
		&p.UserID,
	)
	return err
}
