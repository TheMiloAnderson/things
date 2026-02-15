package models

import (
	"things/internal/db"
)

type Area struct {
	db.Connection
	ID     int
	Name   string
	UserID int
}

func (a *Area) GetById(id int) error {
	row := a.QueryRow("SELECT * FROM areas WHERE id = ?", id)
	err := row.Scan(&a.ID, &a.Name, &a.UserID)
	return err
}

func (a *Area) Save() (int64, error) {
	result, err := a.Exec(
		`INSERT INTO areas (name, user_id)
		VALUES (?, ?)`, a.Name, a.UserID,
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

func (a *Area) Update() error {
	_, err := a.Exec(`UPDATE areas SET name = ? WHERE id = ?`, a.Name, a.ID)
	return err
}

func (a *Area) Delete() error {
	_, err := a.Exec(`DELETE FROM areas WHERE id = ?`, a.ID)
	return err
}

func (a *Area) AllForUser(userID int) ([]Area, error) {
	rows, err := a.Query("SELECT id, name, user_id FROM areas WHERE user_id = ? ORDER BY name ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var areas []Area
	for rows.Next() {
		var ar Area
		err := rows.Scan(&ar.ID, &ar.Name, &ar.UserID)
		if err != nil {
			return nil, err
		}
		areas = append(areas, ar)
	}
	return areas, nil
}
