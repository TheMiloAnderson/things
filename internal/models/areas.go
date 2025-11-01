package models

import (
	"log"
	"things/internal/db"
)

type Area struct {
	db.Table
	ID      int
	Name    string
	User_ID int
}

func (a *Area) GetById(id int) {
	row := a.DB.QueryRow("SELECT * FROM areas WHERE id = ?", id)
	if err := row.Scan(&a.ID, &a.Name, &a.User_ID); err != nil {
		log.Fatal(err)
	}
}
