package api

import (
	"errors"
	"log"
	"net/http"
	"repgen/controller"
	"repgen/web"
	"time"
)

var nilTime = (time.Time{}).UnixNano()
var ReportIntervalDateFormatMap = map[int]string{
	controller.ReportIntervalMonthly: "2006-01",
	controller.ReportIntervalWeekly:  "2006-01-02",
	controller.ReportIntervalDaily:   "2006-01-02",
	controller.ReportIntervalHourly:  "2006-01-02 15",
}

type SubmitReportInput struct {
	Token string                 `json:"token"`
	Date  string                 `json:"date"`
	Data  map[string]interface{} `json:"data"`
}

func SubmitReportHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Parse input
		var submitReportInput SubmitReportInput
		err := web.ParsePostBody(w, r, &submitReportInput)
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
			web.SendJsonResponse(w, response, response.Status)
			return
		}
		// Fetch date format with respect to report interval
		dateFormat, ok := ReportIntervalDateFormatMap[report.Interval]
		if !ok {
			log.Printf("{SubmitReportHandler} ERR: Report id %d has invalid interval %d\n", report.Id, report.Interval)
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Parse date
		date, err := time.Parse(dateFormat, submitReportInput.Date)
		if err != nil {
			var parseError *time.ParseError
			if errors.As(err, &parseError) {
				response := web.Response{Status: http.StatusBadRequest, Message: "Invalid date."}
				web.SendJsonResponse(w, response, response.Status)
				return
			} else {
				log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
				web.SendHttpMethod(w, http.StatusInternalServerError)
				return
			}
		}
		// Check if date is valid
		if date.UnixNano() == nilTime {
			response := web.Response{Status: http.StatusBadRequest, Message: "Invalid date."}
			web.SendJsonResponse(w, response, response.Status)
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
	if len(submitReportInput.Token) != controller.ReportTokenLength*2 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Invalid field length: token"}
	}
	// <date>
	if len(submitReportInput.Date) == 0 {
		return &web.Response{Status: http.StatusBadRequest, Message: "Field cannot be empty: date"}
	}
	// <data>

	return nil
}
