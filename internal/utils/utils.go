package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type HTTPError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func NewHTTPError(statusCode int, message string) HTTPError {
	return HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e HTTPError) Error() string {
	return e.Message
}

// WriteJSONResponse writes a JSON response with the given data and status code.
func WriteJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteErrorResponse writes an error response with the given error message and status code.
func WriteErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	statusCode := http.StatusInternalServerError

	if ve, ok := err.(*validator.ValidationErrors); ok {
		statusCode = http.StatusBadRequest
		translations := ve.Translate(nil) // Translate validation errors if needed
		var errMessages string
		for field, message := range translations {
			errMessages += field + ": " + message + "; "
		}
		err = errors.New(errMessages)
	} else if httpErr, ok := err.(HTTPError); ok {
		statusCode = httpErr.StatusCode
	}

	w.WriteHeader(statusCode)
	response := map[string]string{"error": err.Error()}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
// DecodeJSON decodes a JSON request body into the provided struct.
func DecodeJSON(body io.Reader, v interface{}) error {
	return json.NewDecoder(body).Decode(v)
}
