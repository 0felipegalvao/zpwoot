package validator

import (
	"encoding/json"
	"net/http"
)

// ValidationErrorResponse represents the HTTP response for validation errors
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// WriteValidationError writes a validation error response to the HTTP response writer
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

// WriteValidationErrorWithMessage writes a validation error with a custom message
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

// WriteFieldError writes a single field validation error
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

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(interface{ ValidationErrors() []ValidationError })
	return ok
}

