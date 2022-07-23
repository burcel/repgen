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
	"time"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse input
		var loginInput LoginInput
		err := web.ParsePostBody(w, r, &loginInput)
		if err != nil {
			log.Printf("{LoginHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = loginInputParser(loginInput)
		if err != nil {
			log.Printf("{LoginHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Fetch user by email
		user, err := controller.GetUserByEmail(loginInput.Email)
		if err != nil {
			log.Printf("{LoginHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		} else if user == nil {
			response := web.Response{Message: "Invalid email/password."}
			web.SendJsonResponse(w, response, http.StatusNotFound)
			return
		}
		// Check password
		match, err := security.ComparePasswordAndHash(loginInput.Password, user.Password)
		if err != nil {
			log.Printf("{LoginHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		} else if !match {
			response := web.Response{Message: "Invalid email/password."}
			web.SendJsonResponse(w, response, http.StatusNotFound)
			return
		} else {
			// User & password is correct -> Proceed to session creation

			// Parse session token from cookie
			userSessionCookie, err := web.ParseCookieSessionOptional(r)
			if err != nil {
				log.Printf("{LoginHandler} ERR: %s\n", err.Error())
				var response *web.Response
				if errors.As(err, &response) {
					web.SendJsonResponse(w, response, response.Status)
				} else {
					web.SendHttpMethod(w, http.StatusInternalServerError)
				}
				return
			}
			// Check if token exists
			if userSessionCookie != nil {
				// Token is valid; user is already logged in -> No need to create a new one
				response := web.Response{Status: http.StatusOK, Message: "User is already logged in."}
				web.SendJsonResponse(w, response, http.StatusOK)
				return
			}
			// Generate session token
			// Session duplicate control is skipped here -> Saved 1 query
			session, err := security.GenerateRandomHex(web.CookieSessionLength)
			if err != nil {
				log.Printf("{LoginHandler} ERR: %s\n", err.Error())
				web.SendHttpMethod(w, http.StatusInternalServerError)
				return
			}
			// Register session to database with respect to user id
			userSession := controller.UserSession{UserId: user.Id, Session: session, Created: time.Now().UTC()}
			err = controller.CreateUserSession(userSession)
			if err != nil {
				log.Printf("{LoginHandler} ERR: %s\n", err.Error())
				web.SendHttpMethod(w, http.StatusInternalServerError)
				return
			}
			// Append session to cookie
			cookie := &http.Cookie{
				Name:  web.CookieKeySession,
				Value: session,
				Path:  "/",
				// HttpOnly: true,
			}
			http.SetCookie(w, cookie)
			response := web.Response{Status: http.StatusOK, Message: "User is logged in."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func loginInputParser(loginInput LoginInput) error {
	// <email>
	if len(loginInput.Email) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: email"}
	}
	if len(loginInput.Email) > controller.UserEmailMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: email, max length: %d", controller.UserEmailMaxLength),
		}
	}
	_, err := mail.ParseAddress(loginInput.Email)
	if err != nil {
		return &web.Response{Status: http.StatusBadRequest, Message: "Email is not valid."}
	}
	// <password>
	if len(loginInput.Password) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: password"}
	}
	if len(loginInput.Password) > controller.UserPasswordMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: password, max length: %d", controller.UserPasswordMaxLength),
		}
	}
	return nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		userSession, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{LogoutHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Delete session from database
		err = controller.DeleteUserSession(userSession.Id)
		if err != nil {
			log.Printf("{LogoutHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Reset cookie
		cookie := &http.Cookie{
			Name:    web.CookieKeySession,
			Path:    "/",
			MaxAge:  -1,
			Expires: time.Now().Add(-100 * time.Hour),
			// HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		response := web.Response{Status: http.StatusOK, Message: "User is logged out."}
		web.SendJsonResponse(w, response, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		userSession, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{LogoutHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Delete all user sessions from database
		err = controller.DeleteAllUserSessions(userSession.UserId)
		if err != nil {
			log.Printf("{LogoutHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Reset cookie
		cookie := &http.Cookie{
			Name:    web.CookieKeySession,
			Path:    "/",
			MaxAge:  -1,
			Expires: time.Now().Add(-100 * time.Hour),
			// HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		response := web.Response{Status: http.StatusOK, Message: "User is logged out from everywhere."}
		web.SendJsonResponse(w, response, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}
