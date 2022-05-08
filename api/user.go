package api

import (
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"repgen/controller"
	"repgen/security"
	"repgen/web"
	"strings"
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
		// Parse input
		var userInput UserCreateInput
		err := web.ParsePostBody(w, r, &userInput)
		if err != nil {
			log.Printf("{UserCreateHandler} ERR: %s\n", err.Error())
			return
		}
		// Validate email
		maxEmailLength := 100
		if len(userInput.Email) == 0 {
			response := web.Response{Message: "Field cannot be empty: email"}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		if len(userInput.Email) > maxEmailLength {
			response := web.Response{Message: fmt.Sprintf("Field is too long: email, max length: %d", maxEmailLength)}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		_, err = mail.ParseAddress(userInput.Email)
		if err != nil {
			response := web.Response{Message: "Email is not valid."}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		// Validate password
		maxPasswordLength := 20
		if len(userInput.Password) == 0 {
			response := web.Response{Message: "Field cannot be empty: password"}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		if len(userInput.Password) > maxPasswordLength {
			response := web.Response{Message: fmt.Sprintf("Field is too long: password, max length: %d", maxPasswordLength)}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		// Validate name
		maxNameLength := 100
		if len(userInput.Name) == 0 {
			response := web.Response{Message: "Field cannot be empty: name"}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		if len(userInput.Name) > maxNameLength {
			response := web.Response{Message: fmt.Sprintf("Field is too long: name, max length: %d", maxNameLength)}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		// Hash user password
		hashedPassword, err := security.GenerateHashFromPassword(userInput.Password)
		if err != nil {
			log.Printf("{UserCreateHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Register user
		user := controller.Users{Email: userInput.Email, Password: hashedPassword, Name: userInput.Name, Created: time.Now().UTC()}
		err = controller.CreateUsers(&user)
		if err != nil {
			log.Printf("{UserCreateHandler} ERR: %s\n", err.Error())
			// Check uniqueness of the email
			if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				response := web.Response{Message: "Email already exists."}
				web.SendJsonResponse(w, response, http.StatusNotAcceptable)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
		} else {
			response := web.Response{Status: http.StatusOK, Message: "User is created."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}
