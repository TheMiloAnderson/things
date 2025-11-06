package models

import (
	"things/internal/db"
)

type User struct {
	db.Table
	ID           int
	Name         string
	PasswordHash string
	Inbox        string
}

func (u *User) GetById(id int) error {
	row := u.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&u.ID, &u.Name, &u.PasswordHash, &u.Inbox)
	return err
}

// TODO: worry about this when authN is set up
