package validator

import (
	"encoding/json"
	"net/http"
)

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

func WriteValidationError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errors := FormatValidationErrors(err)

	response := ValidationErrorResponse{
		Error:   "validation_error",
		Message: "Request validation failed",
		Errors:  errors,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func WriteValidationErrorWithMessage(w http.ResponseWriter, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errors := FormatValidationErrors(err)

	response := ValidationErrorResponse{
		Error:   "validation_error",
		Message: message,
		Errors:  errors,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func WriteFieldError(w http.ResponseWriter, field, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := ValidationErrorResponse{
		Error:   "validation_error",
		Message: "Request validation failed",
		Errors: []ValidationError{
			{
				Field:   field,
				Message: message,
			},
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

func IsValidationError(err error) bool {
	_, ok := err.(interface{ ValidationErrors() []ValidationError })
	return ok
}
