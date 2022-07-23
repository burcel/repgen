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

type ProjectCreateInput struct {
	Name string `json:"name"`
}

func ProjectCreateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		userSession, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{ProjectCreateHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var projectCreateInput ProjectCreateInput
		err = web.ParsePostBody(w, r, &projectCreateInput)
		if err != nil {
			log.Printf("{ProjectCreateHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = projectCreateParser(projectCreateInput)
		if err != nil {
			log.Printf("{ProjectCreateHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Register project
		project := controller.Project{Name: projectCreateInput.Name, Created: time.Now().UTC(), CreatedUserId: userSession.UserId}
		err = controller.CreateProject(&project)
		if err != nil {
			log.Printf("{ProjectCreateHandler} ERR: %s\n", err.Error())
			// Check uniqueness of the name
			if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				response := web.Response{Message: "Project name already exists."}
				web.SendJsonResponse(w, response, http.StatusNotAcceptable)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
		} else {
			response := web.Response{Status: http.StatusOK, Message: "Project is created."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func projectCreateParser(projectCreateInput ProjectCreateInput) error {
	// <name>
	if len(projectCreateInput.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(projectCreateInput.Name) > controller.ProjectNameMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", controller.ProjectNameMaxLength),
		}
	}
	return nil
}

type ProjectEditInput struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func ProjectEditHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		_, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{ProjectEditHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var projectEditInput ProjectEditInput
		err = web.ParsePostBody(w, r, &projectEditInput)
		if err != nil {
			log.Printf("{ProjectEditHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = projectEditParser(projectEditInput)
		if err != nil {
			log.Printf("{ProjectEditHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Edit project
		project := controller.Project{Id: projectEditInput.Id, Name: projectEditInput.Name}
		rows, err := controller.UpdateProject(&project)
		if err != nil {
			log.Printf("{ProjectEditHandler} ERR: %s\n", err.Error())
			// Check uniqueness of the name
			if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				response := web.Response{Message: "Project name already exists."}
				web.SendJsonResponse(w, response, http.StatusNotAcceptable)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
		} else if rows != 1 {
			response := web.Response{Message: "Invalid project id."}
			web.SendJsonResponse(w, response, http.StatusInternalServerError)
		} else {
			response := web.Response{Message: "Project is updated."}
			web.SendJsonResponse(w, response, http.StatusOK)
		}
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func projectEditParser(projectEditInput ProjectEditInput) error {
	// <id>
	if projectEditInput.Id < 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be lower than zero: id"}
	}
	// <name>
	if len(projectEditInput.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(projectEditInput.Name) > controller.ProjectNameMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", controller.ProjectNameMaxLength),
		}
	}
	return nil
}

type ProjectSelectInput struct {
	Page int `json:"page"`
}

func ProjectSelectHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		_, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{ProjectSelectHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var projectSelectInput ProjectSelectInput
		err = web.ParsePostBody(w, r, &projectSelectInput)
		if err != nil {
			log.Printf("{ProjectSelectHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = projectSelectParser(projectSelectInput)
		if err != nil {
			log.Printf("{ProjectSelectHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Select all projects
		projects, err := controller.SelectProject(projectSelectInput.Page)
		if err != nil {
			log.Printf("{ProjectSelectHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		web.SendJsonResponse(w, projects, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func projectSelectParser(projectSelectInput ProjectSelectInput) error {
	// <page>
	if projectSelectInput.Page < 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be lower than zero: page"}
	}
	return nil
}
