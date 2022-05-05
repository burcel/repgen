package main

import (
	"log"
	"net/http"
	"repgen/api"
	"repgen/core"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	// Initialize config file
	core.InitializeConfig()
	// Initialize database
	core.InitializeDatabase()

	mux := http.NewServeMux()
	mux.HandleFunc("/login", api.LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":80", mux)
}
