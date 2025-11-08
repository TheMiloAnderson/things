package main

import (
	"fmt"
	"things/internal/models"
)

var dbName = "tasks"

func main() {
	a := models.Area{}
	a.Connect(dbName)
	a.GetById(1)
	fmt.Println(a)
}
