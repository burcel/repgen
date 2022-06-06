package api

import (
	"errors"
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
		// Input validation
		err = UserCreateInputParser(userInput)
		if err != nil {
			log.Printf("{UserCreateHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
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

func UserCreateInputParser(userInput UserCreateInput) error {
	// <email>
	maxEmailLength := 100
	if len(userInput.Email) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: email"}
	}
	if len(userInput.Email) > maxEmailLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: email, max length: %d", maxEmailLength),
		}
	}
	_, err := mail.ParseAddress(userInput.Email)
	if err != nil {
		return &web.Response{Status: http.StatusBadRequest, Message: "Email is not valid."}
	}
	// <password>
	maxPasswordLength := 20
	if len(userInput.Password) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: password"}
	}
	if len(userInput.Password) > maxPasswordLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: password, max length: %d", maxPasswordLength),
		}
	}
	// <name>
	maxNameLength := 100
	if len(userInput.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(userInput.Name) > maxNameLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", maxPasswordLength),
		}
	}
	return nil
}
