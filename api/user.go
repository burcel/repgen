package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/core"
	"time"
)

type UserCreateInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func UserCreateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var userInput UserCreateInput
		err := core.ParsePostBody(w, r, &userInput)
		if err != nil {
			log.Println(err.Error())
			return
		}

		fmt.Printf("%+v\n", userInput)

		user := controller.Users{Email: userInput.Email, Password: userInput.Password, Name: userInput.Name, Created: time.Now().UTC()}
		err = controller.CreateUsers(&user)
		if err != nil {
			log.Println(err.Error())
			return
		} else {
			fmt.Printf("id: %d\n", user.Id)
		}

		responseString, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		core.SendResponse(w, responseString, http.StatusOK)
	default:
		core.SendMethodNotAllowed(w)
	}
}
