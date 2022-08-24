package main

import (
	"log"
	"main/internal/handlers"
	"net/http"
)

func main() {
	mux := routes()

	log.Println("starting channel listener")
	go handlers.ListenToWSChannel()

	log.Println("Starting web saba")

	_ = http.ListenAndServe(":8080", mux)
}
