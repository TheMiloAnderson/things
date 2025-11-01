package main

import (
	"fmt"
	"things/internal/models"
)

var dbName = "tasks"

func main() {
	a := models.Area{}
	a.DBName = dbName
	a.Connect()
	a.GetById(1)
	fmt.Println(a)
}
