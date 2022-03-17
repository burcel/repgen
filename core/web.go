package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Response struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
}

func (response *Response) Error() string {
	return response.Message
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Content type check
	contentType := r.Header["Content-Type"]
	if len(contentType) != 1 || contentType[0] != "application/json" {
		msg := "Content-Type header is not application/json"
		return &Response{Status: http.StatusUnsupportedMediaType, Message: msg}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &Response{Status: http.StatusRequestEntityTooLarge, Message: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &Response{Status: http.StatusBadRequest, Message: msg}
	}

	return nil
}

func ParsePostBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := DecodeJSONBody(w, r, dst)
	if err != nil {
		var errorResponse *Response
		if errors.As(err, &errorResponse) {
			responseBytes, err := json.Marshal(errorResponse)
			if err != nil {
				SendInternalServerError(w)
			} else {
				SendResponse(w, responseBytes, errorResponse.Status)
			}
		} else {
			SendInternalServerError(w)
		}
	}
	return err
}

func SendResponse(w http.ResponseWriter, response []byte, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	w.Write(response)
}

func SendMethodNotAllowed(w http.ResponseWriter) {
	response := Response{Status: http.StatusMethodNotAllowed, Message: http.StatusText(http.StatusMethodNotAllowed)}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	SendResponse(w, responseBytes, http.StatusMethodNotAllowed)
}

func SendInternalServerError(w http.ResponseWriter) {
	response := Response{Status: http.StatusInternalServerError, Message: http.StatusText(http.StatusMethodNotAllowed)}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	SendResponse(w, responseBytes, http.StatusMethodNotAllowed)
}
