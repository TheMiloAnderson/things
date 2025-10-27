package main

import (
	"fmt"
	"things/internal/models"
)

func main() {
	a := models.Area{}
	a.Connect()
	a.GetById(1)
	fmt.Println(a)
}
