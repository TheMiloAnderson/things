package models

import (
	"fmt"
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
		return 0, fmt.Errorf("Save Context: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Save Context: %v", err)
	}
	return id, nil
}

func (c *Context) Update() error {
	_, err := c.Exec(`UPDATE contexts SET name = ? WHERE id = ?`, c.Name, c.ID)
	if err != nil {
		return fmt.Errorf("Update Context: %v", err)
	}
	return nil
}

func (c *Context) Delete() error {
	_, err := c.Exec(`DELETE FROM contexts WHERE id = ?`, c.ID)
	if err != nil {
		return fmt.Errorf("Delete Context: %v", err)
	}
	return nil
}
