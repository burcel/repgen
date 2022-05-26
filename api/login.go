package api

import (
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
		// Validate email
		maxEmailLength := 100
		if len(loginInput.Email) == 0 {
			response := web.Response{Message: "Field cannot be empty: email"}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		if len(loginInput.Email) > maxEmailLength {
			response := web.Response{Message: fmt.Sprintf("Field is too long: email, max length: %d", maxEmailLength)}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		_, err = mail.ParseAddress(loginInput.Email)
		if err != nil {
			response := web.Response{Message: "Email is not valid."}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		// Validate password
		maxPasswordLength := 20
		if len(loginInput.Password) == 0 {
			response := web.Response{Message: "Field cannot be empty: password"}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		if len(loginInput.Password) > maxPasswordLength {
			response := web.Response{Message: fmt.Sprintf("Field is too long: password, max length: %d", maxPasswordLength)}
			web.SendJsonResponse(w, response, http.StatusBadRequest)
			return
		}
		// Fetch user by email
		user, err := controller.GetUsersByEmail(loginInput.Email)
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
			sessionCookie, err := r.Cookie(web.CookieKeySession)
			// If there exists a cookie -> Check validity
			if sessionCookie != nil {
				userSession, err := controller.GetUserSession(sessionCookie.Value)
				if err != nil {
					log.Printf("{LoginHandler} ERR: %s\n", err.Error())
					web.SendHttpMethod(w, http.StatusInternalServerError)
					return
				}
				if userSession == nil {
					// Token does not exist in database
					response := web.Response{Message: "Invalid authentication!"}
					web.SendJsonResponse(w, response, http.StatusUnauthorized)
					return
				} else {
					// Token exists, no need to create a new one
					response := web.Response{Status: http.StatusOK, Message: "User is already logged in."}
					web.SendJsonResponse(w, response, http.StatusOK)
					return
				}
			}
			// Generate session token
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		sessionCookie, err := r.Cookie(web.CookieKeySession)
		if err != nil {
			// Token does not exist in cookie
			response := web.Response{Message: "Invalid authentication!"}
			web.SendJsonResponse(w, response, http.StatusUnauthorized)
			return
		}
		// If there exists a cookie -> Check validity
		if sessionCookie != nil {
			userSession, err := controller.GetUserSession(sessionCookie.Value)
			if err != nil {
				log.Printf("{LogoutHandler} ERR: %s\n", err.Error())
				web.SendHttpMethod(w, http.StatusInternalServerError)
				return
			}
			if userSession == nil {
				// Token does not exist in database
				response := web.Response{Message: "Invalid authentication!"}
				web.SendJsonResponse(w, response, http.StatusUnauthorized)
				return
			} else {
				// Token exists -> Proceed to logout
				err = controller.DeleteUserSession(userSession.Id)
				if err != nil {
					log.Printf("{LogoutHandler} ERR: %s\n", err.Error())
					web.SendHttpMethod(w, http.StatusInternalServerError)
					return
				}
				// Reset cookie
				cookie := &http.Cookie{
					Name:    web.CookieKeySession,
					MaxAge:  -1,
					Expires: time.Now().Add(-100 * time.Hour),
					// HttpOnly: true,
				}
				http.SetCookie(w, cookie)
			}
		}
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}
