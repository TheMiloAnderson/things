package models

import (
	"fmt"
	"things/internal/db"
)

type Area struct {
	db.Table
	ID     int
	Name   string
	UserID int
}

func (a *Area) GetById(id int) error {
	row := a.QueryRow("SELECT * FROM areas WHERE id = ?", id)
	err := row.Scan(&a.ID, &a.Name, &a.UserID)
	if err != nil {
		return fmt.Errorf("Update Area: %v", err)
	}
	return nil
}

func (a *Area) Save() (int64, error) {
	result, err := a.Exec(
		`INSERT INTO areas (name, user_id)
		VALUES (?, ?)`, a.Name, a.UserID,
	)
	if err != nil {
		return 0, fmt.Errorf("Save Area: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Save Area: %v", err)
	}
	return id, nil
}

func (a *Area) Update() error {
	_, err := a.Exec(`UPDATE areas SET name = ? WHERE id = ?`, a.Name, a.ID)
	if err != nil {
		return fmt.Errorf("Update Area: %v", err)
	}
	return nil
}

func (a *Area) Delete() error {
	_, err := a.Exec(`DELETE FROM areas WHERE id = ?`, a.ID)
	if err != nil {
		return fmt.Errorf("Delete Area: %v", err)
	}
	return nil
}
