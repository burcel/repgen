package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/web"
	"strings"
	"time"
)

type ReportCreateInput struct {
	ProjectId   int    `json:"project_id"`
	Name        string `json:"name"`
	Interval    int    `json:"interval"`
	Description string `json:"description"`
}

func ReportCreateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		userSession, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var reportCreateInput ReportCreateInput
		err = web.ParsePostBody(w, r, &reportCreateInput)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = reportCreateParser(reportCreateInput)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Register project
		report := controller.Report{
			ProjectId:     reportCreateInput.ProjectId,
			Name:          reportCreateInput.Name,
			Interval:      reportCreateInput.Interval,
			Description:   reportCreateInput.Description,
			Created:       time.Now().UTC(),
			CreatedUserId: userSession.UserId,
		}
		err = controller.CreateReport(&report)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			// Check uniqueness of the name
			if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				response := web.Response{Message: "Project name already exists."}
				web.SendJsonResponse(w, response, http.StatusNotAcceptable)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
		} else {
			response := web.Response{Status: http.StatusOK, Message: "Report is created."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func reportCreateParser(reportCreateInput ReportCreateInput) error {
	// <name>
	maxNameLength := 100
	if len(reportCreateInput.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(reportCreateInput.Name) > maxNameLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", maxNameLength),
		}
	}
	// <interval>
	if reportCreateInput.Interval != 1 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field is invalid: interval"}
	}
	// <description>
	maxDescriptionLength := 400
	if len(reportCreateInput.Description) > maxDescriptionLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: description, max length: %d", maxNameLength),
		}
	}
	return nil
}
