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
	mux.HandleFunc("/logout", api.LogoutHandler)
	mux.HandleFunc("/logout/all", api.LogoutAllHandler)
	mux.HandleFunc("/user/create", api.UserCreateHandler)
	mux.HandleFunc("/user/edit", api.UserEditHandler)
	mux.HandleFunc("/user/password", api.UserChangePasswordHandler)
	mux.HandleFunc("/project/create", api.ProjectCreateHandler)
	mux.HandleFunc("/project/edit", api.ProjectEditHandler)
	mux.HandleFunc("/project/", api.ProjectSelectHandler)
	mux.HandleFunc("/report/create", api.ReportCreateHandler)
	mux.HandleFunc("/report/refresh", api.ReportRefreshTokenHandler)
	mux.HandleFunc("/report/", api.ReportSelectHandler)
	mux.HandleFunc("/submit", api.SubmitReportHandler)

	log.Println("Listening...")
	http.ListenAndServe(":80", mux)
}
