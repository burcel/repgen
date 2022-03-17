package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"repgen/api"
)

type TestStruct struct {
	Test string
}

func parseGhPost(rw http.ResponseWriter, request *http.Request) {
	var t TestStruct
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&t)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", t.Test)

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", api.LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":80", mux)
}
