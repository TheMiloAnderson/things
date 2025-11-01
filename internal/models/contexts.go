package models

import "things/internal/db"

type Context struct {
	db.Table
	ID     int
	Name   string
	UserID int
}

func (c *Context) GetById(id int) error {
	row := c.DB.QueryRow("SELECT * FROM contexts WHERE id = ?", id)
	err := row.Scan(&c.ID, &c.Name, &c.UserID)
	return err
}
