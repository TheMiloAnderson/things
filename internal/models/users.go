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

func (u *User) GetByName(name string) error {
	row := u.QueryRow("SELECT * FROM users WHERE name = ?", name)
	err := row.Scan(&u.ID, &u.Name, &u.PasswordHash, &u.Inbox)
	return err
}

func (u *User) Update() error {
	_, err := u.Exec(
		`UPDATE users SET name = ?, password_hash = ?, inbox = ?
		WHERE id = ?`, u.Name, u.PasswordHash, u.Inbox, u.ID,
	)
	return err
}
