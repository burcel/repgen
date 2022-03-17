package api

import (
	"encoding/json"
	"errors"
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
		err := core.DecodeJSONBody(w, r, &loginInput)
		if err != nil {
			var errorResponse *core.Response
			if errors.As(err, &errorResponse) {
				responseBytes, err := json.Marshal(errorResponse)
				if err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				} else {
					core.SendResponse(w, responseBytes, errorResponse.Status)
				}
			} else {
				log.Println(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
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
		response := core.Response{Status: http.StatusMethodNotAllowed, Message: http.StatusText(http.StatusMethodNotAllowed)}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		core.SendResponse(w, responseBytes, http.StatusMethodNotAllowed)
	}
}
