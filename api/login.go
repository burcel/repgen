package api

import (
	"fmt"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/web"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var loginInput LoginInput
		err := web.ParsePostBody(w, r, &loginInput)
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

		response := web.Response{Status: http.StatusOK, Message: "Successfully logged in."}
		web.SendJsonResponse(w, response, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}
