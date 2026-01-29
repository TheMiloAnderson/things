package main

import (
	"fmt"
	"log"
	"net/http"
)

var dbName = "tasks"

type PageData struct {
	IsAuthenticated bool
	Username        string
	Data            any
}

func main() {
	http.HandleFunc("/", AuthRequired(inboxHandler))

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("Server is starting on port 8888...")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
