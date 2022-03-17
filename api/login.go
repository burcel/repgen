package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"repgen/core"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var loginInput LoginInput
		err := core.ParsePostBody(w, r, &loginInput)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Printf("%+v\n", loginInput)

		responseString, err := json.Marshal(loginInput)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		core.SendResponse(w, responseString, http.StatusOK)
	default:
		core.SendMethodNotAllowed(w)
	}
}
