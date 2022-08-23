package main

import (
	"log"
	"net/http"
)

func main() {
	mux := routes()

	log.Println("Starting web saba")

	_ = http.ListenAndServe(":8080", mux)
}
