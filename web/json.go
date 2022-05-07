package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Parse HTTP POST body and send response if any anomaly happens
func ParsePostBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := DecodeJSONBody(w, r, dst)
	if err != nil {
		var errorResponse *Response
		if errors.As(err, &errorResponse) {
			SendJsonResponse(w, errorResponse, errorResponse.Status)
		} else {
			SendHttpMethod(w, http.StatusInternalServerError)
		}
	}
	return err
}

// Decode request body into given struct format
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Content type check
	contentType := r.Header["Content-Type"]
	if len(contentType) != 1 || contentType[0] != "application/json" {
		msg := "Content-Type header is not application/json"
		return &Response{Status: http.StatusUnsupportedMediaType, Message: msg}
	}
	// Set max HTTP body size 1048576 = 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	// Decode HTTP request body
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d).", syntaxError.Offset)
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON."
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field \"")
			fieldName = strings.TrimSuffix(fieldName, "\"")
			msg := fmt.Sprintf("Request body contains unknown field: %s", fieldName)
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty."
			return &Response{Status: http.StatusBadRequest, Message: msg}

		case err.Error() == "http: request body too large":
			// Value set in http.Server.MaxHeaderBytes
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
