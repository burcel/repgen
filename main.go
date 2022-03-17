package main

import (
	"log"
	"net/http"
	"repgen/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", api.LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":80", mux)
}
