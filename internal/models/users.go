package models

import (
	"database/sql"
	"errors"
	"time"

	"things/internal/db"
)

type User struct {
	db.Connection
	ID                int
	Name              string
	Email             string
	EmailVerifiedAt   sql.NullTime
	PasswordHash      string
	PasswordChangedAt sql.NullTime
	Inbox             string
}

func (u *User) scan(row *sql.Row) error {
	return row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.EmailVerifiedAt,
		&u.PasswordHash,
		&u.PasswordChangedAt,
		&u.Inbox,
	)
}

func (u *User) GetById(id int) error {
	row := u.QueryRow(
		`SELECT id, name, email, email_verified_at, password_hash, password_changed_at, inbox
		FROM users WHERE id = ?`, id,
	)
	return u.scan(row)
}

func (u *User) GetByName(name string) error {
	row := u.QueryRow(
		`SELECT id, name, email, email_verified_at, password_hash, password_changed_at, inbox
		FROM users WHERE name = ?`, name,
	)
	return u.scan(row)
}

func (u *User) GetByEmail(email string) error {
	row := u.QueryRow(
		`SELECT id, name, email, email_verified_at, password_hash, password_changed_at, inbox
		FROM users WHERE email = ?`, email,
	)
	return u.scan(row)
}

func (u *User) Create() (int64, error) {
	res, err := u.Exec(
		`INSERT INTO users (name, email, password_hash, inbox)
		VALUES (?, ?, ?, ?)`,
		u.Name, u.Email, u.PasswordHash, u.Inbox,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (u *User) Update() error {
	_, err := u.Exec(
		`UPDATE users SET name = ?, email = ?, password_hash = ?, inbox = ?
		WHERE id = ?`,
		u.Name, u.Email, u.PasswordHash, u.Inbox, u.ID,
	)
	return err
}

func (u *User) MarkEmailVerified() error {
	now := time.Now()
	_, err := u.Exec(
		`UPDATE users SET email_verified_at = ? WHERE id = ?`,
		now, u.ID,
	)
	if err != nil {
		return err
	}
	u.EmailVerifiedAt = sql.NullTime{Time: now, Valid: true}
	return nil
}

func (u *User) UpdatePassword(newHash string) error {
	now := time.Now()
	_, err := u.Exec(
		`UPDATE users SET password_hash = ?, password_changed_at = ? WHERE id = ?`,
		newHash, now, u.ID,
	)
	if err != nil {
		return err
	}
	u.PasswordHash = newHash
	u.PasswordChangedAt = sql.NullTime{Time: now, Valid: true}
	return nil
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt.Valid
}

var ErrDuplicateUser = errors.New("duplicate user")
