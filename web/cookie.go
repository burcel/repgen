package web

import (
	"net/http"
	"strings"

	"repgen/controller"
)

const CookieSessionLength = 32
const CookieKeySession = "session"

func ParseCookieSessionOptional(r *http.Request) (*controller.UserSession, error) {
	// Parse session token from cookie
	sessionCookie, err := r.Cookie(CookieKeySession)
	if err != nil {
		if strings.Contains(err.Error(), "named cookie not present") {
			// Cookie does not exist
			return nil, nil
		}
		return nil, err
	}
	// A cookie exists -> Check validity
	userSession, err := controller.GetUserSession(sessionCookie.Value)
	if err != nil {
		return nil, err
	}
	return userSession, nil
}

func ParseCookieSession(r *http.Request) (*controller.UserSession, error) {
	// Parse session token from cookie
	sessionCookie, err := r.Cookie(CookieKeySession)
	if err != nil || sessionCookie == nil {
		// Token does not exist inside cookie
		response := &Response{Status: http.StatusUnauthorized, Message: "Invalid authentication!"}
		return nil, response
	}
	// A cookie exists -> Check validity
	userSession, err := controller.GetUserSession(sessionCookie.Value)
	if err != nil {
		return nil, err
	}
	if userSession == nil {
		// Token does not exist in database
		response := &Response{Status: http.StatusUnauthorized, Message: "Invalid authentication!"}
		return nil, response
	}
	return userSession, nil
}
