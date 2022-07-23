package api

import (
	"errors"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/web"
)

type SubmitReportInput struct {
	Token string                 `json:"token"`
	Date  string                 `json:"date"`
	Data  map[string]interface{} `json:"data"`
}

func SubmitReportHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse session token from cookie
		_, err := web.ParseCookieSession(r)
		if err != nil {
			log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Parse input
		var submitReportInput SubmitReportInput
		err = web.ParsePostBody(w, r, &submitReportInput)
		if err != nil {
			log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
			return
		}
		// Input validation
		err = SubmitReportParser(submitReportInput)
		if err != nil {
			log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Fetch report from token
		report, err := controller.GetReportByToken(submitReportInput.Token)
		if err != nil {
			log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		if report == nil {
			response := web.Response{Status: http.StatusBadRequest, Message: "Invalid token."}
			web.SendJsonResponse(w, response, http.StatusOK)
			return
		}
		// Fetch report columns
		err = controller.PopulateReportColumns(report)
		if err != nil {
			log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		response := web.Response{Status: http.StatusOK, Message: "Report data is submitted."}
		web.SendJsonResponse(w, response, http.StatusOK)
	default:
		web.SendHttpMethod(w, http.StatusMethodNotAllowed)
	}
}

func SubmitReportParser(submitReportInput SubmitReportInput) error {
	// <token>
	if len(submitReportInput.Token) != controller.ReportTokenLength {
		return &web.Response{Status: http.StatusBadRequest, Message: "Invalid field length: token"}
	}
	// <date>
	// <data>
	return nil
}
