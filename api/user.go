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
		err = userCreateInputParser(userInput)
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
		user := controller.User{Email: userInput.Email, Password: hashedPassword, Name: userInput.Name, Created: time.Now().UTC()}
		err = controller.CreateUser(&user)
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

func userCreateInputParser(userInput UserCreateInput) error {
	// <email>
	if len(userInput.Email) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: email"}
	}
	if len(userInput.Email) > controller.UserEmailMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: email, max length: %d", controller.UserEmailMaxLength),
		}
	}
	_, err := mail.ParseAddress(userInput.Email)
	if err != nil {
		return &web.Response{Status: http.StatusBadRequest, Message: "Email is not valid."}
	}
	// <password>
	if len(userInput.Password) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: password"}
	}
	if len(userInput.Password) > controller.UserPasswordMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: password, max length: %d", controller.UserPasswordMaxLength),
		}
	}
	// <name>
	if len(userInput.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(userInput.Name) > controller.UserNameMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", controller.UserNameMaxLength),
		}
	}
	return nil
}

type UserEditInput struct {
	Name string `json:"name"`
}

func UserEditHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		userSession, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{UserEditHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var userEdit UserEditInput
		err = web.ParsePostBody(w, r, &userEdit)
		if err != nil {
			log.Printf("{UserEditHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = userEditInputParser(userEdit)
		if err != nil {
			log.Printf("{UserEditHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		user := controller.User{Id: userSession.UserId, Name: userEdit.Name}
		// Edit user
		rows, err := controller.UpdateUser(user)
		if err != nil {
			log.Printf("{UserEditHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
		} else if rows != 1 {
			log.Printf("{UserEditHandler} ERR: Update failed for user id: %d\n", user.Id)
			web.SendHttpMethod(w, http.StatusBadRequest)
		} else {
			response := web.Response{Message: "User is updated."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	}
}

func userEditInputParser(userEdit UserEditInput) error {
	// <name>
	if len(userEdit.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(userEdit.Name) > controller.UserNameMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", controller.UserNameMaxLength),
		}
	}
	return nil
}

type UserChangePasswordInput struct {
	Password string `json:"password"`
}

func UserChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		userSession, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{UserChangePasswordHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var userChangePasswordInput UserChangePasswordInput
		err = web.ParsePostBody(w, r, &userChangePasswordInput)
		if err != nil {
			log.Printf("{UserChangePasswordHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = UserChangePasswordParser(userChangePasswordInput)
		if err != nil {
			log.Printf("{UserChangePasswordHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Hash user password
		hashedPassword, err := security.GenerateHashFromPassword(userChangePasswordInput.Password)
		if err != nil {
			log.Printf("{UserChangePasswordHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		user := controller.User{Id: userSession.UserId, Password: hashedPassword}
		// Update user password
		rows, err := controller.UpdateUserPassword(user)
		if err != nil {
			log.Printf("{UserChangePasswordHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
		} else if rows != 1 {
			log.Printf("{UserChangePasswordHandler} ERR: Update failed for user id: %d\n", user.Id)
			web.SendHttpMethod(w, http.StatusBadRequest)
		} else {
			response := web.Response{Message: "User is updated."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	}
}

func UserChangePasswordParser(userChangePasswordInput UserChangePasswordInput) error {
	// <password>
	if len(userChangePasswordInput.Password) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: password"}
	}
	if len(userChangePasswordInput.Password) > controller.UserPasswordMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: password, max length: %d", controller.UserPasswordMaxLength),
		}
	}
	return nil
}
