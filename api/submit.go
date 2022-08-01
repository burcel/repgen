package api

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"reflect"
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
		// Validate columns
		err = SubmitReportColumnParser(report, submitReportInput)
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
		// Record column values

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
	if len(submitReportInput.Data) == 0 {
		return &web.Response{
			Status:  http.StatusBadRequest,
			Message: "Field is empty: data",
		}
	}
	return nil
}

func SubmitReportColumnParser(report *controller.Report, submitReportInput SubmitReportInput) error {
	// Create Map: Column name -> Type
	ReportColumnNameTypeMap := make(map[string]int)
	for _, reportColumn := range report.Columns {
		ReportColumnNameTypeMap[reportColumn.Name] = reportColumn.Type
	}
	// Validate column types
	for columnName, value := range submitReportInput.Data {
		if columnType, ok := ReportColumnNameTypeMap[columnName]; ok {
			switch columnType {
			case controller.ReportColumnTypeStr:
				if reflect.TypeOf(value).String() != "string" {
					return &web.Response{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid column type: %s", columnName),
					}
				}
			case controller.ReportColumnTypeInt:
				if reflect.TypeOf(value).String() != "float64" {
					// Value can be cast to int here
					return &web.Response{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid column type: %s", columnName),
					}
				} else {
					if value != math.Trunc(value.(float64)) {
						return &web.Response{
							Status:  http.StatusBadRequest,
							Message: fmt.Sprintf("Invalid column type: %s", columnName),
						}
					}
				}
			case controller.ReportColumnTypeFloat:
				if reflect.TypeOf(value).String() != "float64" {
					return &web.Response{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid column type: %s", columnName),
					}
				}
			case controller.ReportColumnTypeFormula:
				return &web.Response{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("Data cannot be send to formula: %s", columnName),
				}
			default:
				return &web.Response{
					Status:  http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
				}
			}
		} else {
			return &web.Response{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Column does not exist: %s", columnName),
			}
		}
	}
	return nil
}
