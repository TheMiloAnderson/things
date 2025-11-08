package models

import (
	"things/internal/db"
)

type Context struct {
	db.Table
	ID     int
	Name   string
	UserID int
}

func (c *Context) GetById(id int) error {
	row := c.QueryRow("SELECT * FROM contexts WHERE id = ?", id)
	err := row.Scan(&c.ID, &c.Name, &c.UserID)
	return err
}

func (c *Context) Save() (int64, error) {
	result, err := c.Exec(
		`INSERT INTO contexts (name, user_id)
		VALUES (?, ?)`, c.Name, c.UserID,
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

func (c *Context) Update() error {
	_, err := c.Exec(`UPDATE contexts SET name = ? WHERE id = ?`, c.Name, c.ID)
	return err
}

func (c *Context) Delete() error {
	_, err := c.Exec(`DELETE FROM contexts WHERE id = ?`, c.ID)
	return err
}
