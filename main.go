package main

import (
	"database/sql"
	"log"
	"net/http"
	"repgen/api"
	"repgen/core"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	// Initialize config file
	core.InitializeConfig()
	log.Println(core.Config.Version)
	// Initialize database
	db, err := sql.Open("pgx", "postgres://burakc:burakc@localhost/repgen")
	if err != nil {
		log.Println(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	log.Println("database is reachable")

	mux := http.NewServeMux()
	mux.HandleFunc("/login", api.LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":80", mux)
}
