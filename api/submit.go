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
		err = submitReportParser(submitReportInput)
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
		// Parse report date
		date, err := submitReportDateParser(report, submitReportInput.Date)
		if err != nil {
			var response *web.Response
			if errors.As(err, &response) {
				web.SendJsonResponse(w, response, response.Status)
			} else {
				web.SendHttpMethod(w, http.StatusInternalServerError)
			}
			return
		}
		// Populate report columns
		err = controller.PopulateReportColumns(report)
		if err != nil {
			log.Printf("{SubmitReportHandler} ERR: %s\n", err.Error())
			web.SendHttpMethod(w, http.StatusInternalServerError)
			return
		}
		// Validate columns
		reportColumnIdValueMap, err := submitReportColumnParser(report, submitReportInput)
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
		reportData := controller.ReportData{
			ReportDate: *date,
			SentDate:   time.Now().UTC(),
			ColumnMap:  reportColumnIdValueMap,
		}
		// Insert report data
		err = controller.InsertReportData(&reportData)
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

func submitReportParser(submitReportInput SubmitReportInput) error {
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

func submitReportDateParser(report *controller.Report, submitDate string) (*time.Time, error) {
	// Fetch date format with respect to report interval
	dateFormat, ok := ReportIntervalDateFormatMap[report.Interval]
	if !ok {
		log.Printf("{SubmitReportDateParser} ERR: Report id %d has invalid interval %d\n", report.Id, report.Interval)
		response := &web.Response{
			Status:  http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
		return nil, response
	}
	// Parse date
	date, err := time.Parse(dateFormat, submitDate)
	if err != nil {
		var parseError *time.ParseError
		if errors.As(err, &parseError) {
			response := &web.Response{Status: http.StatusBadRequest, Message: "Invalid date."}
			return nil, response
		} else {
			log.Printf("{SubmitReportDateParser} ERR: %s\n", err.Error())
			response := &web.Response{
				Status:  http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			}
			return nil, response
		}
	}
	// Check if date is valid
	if date.UnixNano() == nilTime {
		response := &web.Response{Status: http.StatusBadRequest, Message: "Invalid date."}
		return nil, response
	}
	return &date, nil
}

func submitReportColumnParser(report *controller.Report, submitReportInput SubmitReportInput) (map[int]interface{}, error) {
	// Map: Column name -> Type
	reportColumnNameTypeMap := make(map[string]int)
	// Map: Column name -> Column id
	reportColumnNameIdMap := make(map[string]int)
	// Map: Column id -> Value
	reportColumnIdValueMap := make(map[int]interface{})
	for _, reportColumn := range report.Columns {
		reportColumnNameTypeMap[reportColumn.Name] = reportColumn.Type
		reportColumnNameIdMap[reportColumn.Name] = reportColumn.Id
	}
	// Validate column types
	for columnName, value := range submitReportInput.Data {
		if columnType, ok := reportColumnNameTypeMap[columnName]; ok {
			switch columnType {
			case controller.ReportColumnTypeStr:
				if reflect.TypeOf(value).String() != "string" {
					response := &web.Response{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid column type: %s", columnName),
					}
					return nil, response
				}
			case controller.ReportColumnTypeInt:
				if reflect.TypeOf(value).String() != "float64" {
					response := &web.Response{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid column type: %s", columnName),
					}
					return nil, response
				} else {
					// Check if value is int or float i.e. 3.0 or 3
					// -> Go serializes integer for empty interface as float64
					if value != math.Trunc(value.(float64)) {
						response := &web.Response{
							Status:  http.StatusBadRequest,
							Message: fmt.Sprintf("Invalid column type: %s", columnName),
						}
						return nil, response
					}
				}
			case controller.ReportColumnTypeFloat:
				if reflect.TypeOf(value).String() != "float64" {
					response := &web.Response{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid column type: %s", columnName),
					}
					return nil, response
				}
			case controller.ReportColumnTypeFormula:
				response := &web.Response{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("Data cannot be send to formula: %s", columnName),
				}
				return nil, response
			default:
				log.Printf("{submitReportColumnParser} ERR: Report id %d has invalid column type for: %s\n", report.Id, columnName)
				response := &web.Response{
					Status:  http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
				}
				return nil, response
			}
			// Add value to map
			columnId := reportColumnNameIdMap[columnName]
			reportColumnIdValueMap[columnId] = value
		} else {
			response := &web.Response{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Column does not exist: %s", columnName),
			}
			return nil, response
		}
	}
	return reportColumnIdValueMap, nil
}
