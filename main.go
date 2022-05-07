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
	// Start server
	mux := http.NewServeMux()
	mux.HandleFunc("/login", api.LoginHandler)
	mux.HandleFunc("/user/create", api.UserCreateHandler)
	log.Println("Listening...")
	http.ListenAndServe(":80", mux)
}
