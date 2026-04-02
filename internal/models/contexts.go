package models

import (
	"things/internal/db"
)

type Context struct {
	db.Connection
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

// AllForUser returns all contexts for the user, ordered by name.
func (c *Context) AllForUser(userID int) ([]Context, error) {
	rows, err := c.Query(`SELECT id, name, user_id FROM contexts WHERE user_id = ? ORDER BY name ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contexts []Context
	for rows.Next() {
		var ctx Context
		if err := rows.Scan(&ctx.ID, &ctx.Name, &ctx.UserID); err != nil {
			return nil, err
		}
		contexts = append(contexts, ctx)
	}
	return contexts, rows.Err()
}
