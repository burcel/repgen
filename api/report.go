package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/security"
	"repgen/web"
	"time"
)

type ReportCreateInput struct {
	ProjectId   int                     `json:"project_id"`
	Name        string                  `json:"name"`
	Interval    int                     `json:"interval"`
	Description string                  `json:"description"`
	Definition  []ReportDefinitionInput `json:"definition"`
}

type ReportDefinitionInput struct {
	Name    string `json:"name"`
	Type    int    `json:"type"`
	Formula string `json:"formula"`
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
		// Generate token
		var token string
		for {
			token, err = security.GenerateRandomHex(controller.ReportTokenLength)
			if err != nil {
				log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
				web.SendHttpMethod(w, http.StatusInternalServerError)
				return
			}
			tokenReport, err := controller.GetReportByToken(token)
			if err != nil {
				log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
				web.SendHttpMethod(w, http.StatusInternalServerError)
				return
			}
			if tokenReport == nil {
				// Token does not exist in database
				break
			}
		}
		// Create report
		report := controller.Report{
			ProjectId:     reportCreateInput.ProjectId,
			Name:          reportCreateInput.Name,
			Interval:      reportCreateInput.Interval,
			Token:         token,
			Description:   reportCreateInput.Description,
			Created:       time.Now().UTC(),
			CreatedUserId: userSession.UserId,
		}
		// Register report
		err = controller.CreateReport(&report)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Create column definitions
		report.Columns = make([]controller.ReportColumn, len(reportCreateInput.Definition))
		for index, column := range reportCreateInput.Definition {
			report.Columns[index] = controller.ReportColumn{
				ReportId:      report.Id,
				Name:          column.Name,
				Type:          column.Type,
				Formula:       column.Formula,
				Created:       time.Now().UTC(),
				CreatedUserId: userSession.UserId,
			}
		}
		// Register column definitions
		err = controller.CreateReportColumns(report.Columns)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Create report data table with respect to columns
		err = controller.CreateReportData(report)
		if err != nil {
			log.Printf("{ReportCreateHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}

		response := web.Response{Status: http.StatusOK, Message: "Report is created."}
		web.SendJsonResponse(w, response, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func reportCreateParser(reportCreateInput ReportCreateInput) error {
	// <name>
	if len(reportCreateInput.Name) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: name"}
	}
	if len(reportCreateInput.Name) > controller.ReportNameMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: name, max length: %d", controller.ReportNameMaxLength),
		}
	}
	// <interval>
	if _, ok := controller.ReportIntervalMap[reportCreateInput.Interval]; !ok {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field is invalid: interval"}
	}
	// <description>
	if len(reportCreateInput.Description) > controller.ReportDescriptionMaxLength {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too long: description, max length: %d", controller.ReportDescriptionMaxLength),
		}
	}
	// <definition>
	if len(reportCreateInput.Definition) == 0 {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: "Field is empty: definition",
		}
	}
	if len(reportCreateInput.Definition) > controller.ReportColumnMaxCount {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Field is too many: definition, max count: %d", controller.ReportColumnMaxCount),
		}
	}
	var emptyStruct struct{}
	columnNameMap := make(map[string]struct{})
	for index, column := range reportCreateInput.Definition {
		// Column name
		if len(column.Name) == 0 {
			return &web.Response{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Field cannot be empty: name at index %d", index+1),
			}
		}
		if len(column.Name) > controller.ReportColumnNameMaxLength {
			return &web.Response{
				Status: http.StatusBadRequest,
				Message: fmt.Sprintf("Field is too long: name at index %d, max length: %d",
					index+1, controller.ReportColumnNameMaxLength),
			}
		}
		// Duplicate control
		if _, ok := columnNameMap[column.Name]; ok {
			return &web.Response{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Duplicate column name at index %d", index+1),
			}
		}
		columnNameMap[column.Name] = emptyStruct
		// Column type
		if _, ok := controller.ReportColumnTypeMap[column.Type]; !ok {
			return &web.Response{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Field is invalid: type at index %d", index+1),
			}
		}
		// Column type -> Formula
		if column.Type == controller.ReportColumnTypeFormula {
			if len(column.Formula) == 0 {
				return &web.Response{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("Field cannot be empty: formula at index %d", index+1),
				}
			}
			if len(column.Formula) > controller.ReportColumnFormulaMaxLength {
				return &web.Response{
					Status: http.StatusBadRequest,
					Message: fmt.Sprintf("Field is too long: formula at index %d, max length: %d",
						index+1, controller.ReportColumnFormulaMaxLength),
				}
			}
		}
	}

	return nil
}

type ReportSelectInput struct {
	ProjectId int `json:"project_id"`
	Page      int `json:"page"`
}

func ReportSelectHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		_, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{ReportSelectHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var reportSelectInput ReportSelectInput
		err = web.ParsePostBody(w, r, &reportSelectInput)
		if err != nil {
			log.Printf("{ReportSelectHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = reportSelectParser(reportSelectInput)
		if err != nil {
			log.Printf("{ReportSelectHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Select all projects
		projects, err := controller.SelectReport(reportSelectInput.ProjectId, reportSelectInput.Page)
		if err != nil {
			log.Printf("{ReportSelectHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		web.SendJsonResponse(w, projects, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func reportSelectParser(reportSelectInput ReportSelectInput) error {
	// <page>
	if reportSelectInput.Page < 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be lower than zero: page"}
	}
	return nil
}
