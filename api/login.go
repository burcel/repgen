package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/core"
)

type LoginInput struct {
	Email    string `json:"email"`
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

		var user *controller.Users
		user, err = controller.GetUsersByEmail(loginInput.Email)
		if err != nil {
			log.Println(err.Error())
			return
		} else if user == nil {
			log.Println("No user")
		} else {
			fmt.Printf("%s\n", user.Email)
		}

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
