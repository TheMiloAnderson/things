package models

import (
	"things/internal/db"
)

type Area struct {
	db.Table
	ID      int
	Name    string
	User_ID int
}

func (a *Area) GetById(id int) error {
	row := a.DB.QueryRow("SELECT * FROM areas WHERE id = ?", id)
	err := row.Scan(&a.ID, &a.Name, &a.User_ID)
	return err
}
