package web

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
}

func (response *Response) Error() string {
	return response.Message
}

// Send the given byte array and HTTP status as response to the ResponseWriter
func SendResponse(w http.ResponseWriter, response []byte, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	w.Write(response)
}

// Send given HTTP method to the ResponseWriter,
// If the response cannot be serialized; send HTTP internal server error (500)
func SendHttpMethod(w http.ResponseWriter, httpStatus int) {
	response := Response{Message: http.StatusText(httpStatus)}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	SendResponse(w, responseBytes, httpStatus)
}

// Turn given struct to bytes and send it with HTTP status
func SendJsonResponse(w http.ResponseWriter, response interface{}, httpStatus int) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		SendHttpMethod(w, http.StatusInternalServerError)
	} else {
		SendResponse(w, responseBytes, httpStatus)
	}
}
