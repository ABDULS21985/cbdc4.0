package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse follows the design doc format
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
}

// SuccessResponse is a generic success wrapper (optional, but good for consistency)
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// WriteError writes a standardized JSON error response
func WriteError(w http.ResponseWriter, statusCode int, code, message, traceID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		TraceID: traceID,
	})
}

// WriteSuccess writes a standardized JSON success response
func WriteSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// If data is already a struct/map, encode it directly or wrap it?
	// Design doc doesn't strictly enforce a "data" wrapper for success, but it's good practice.
	// For now, to minimize friction with existing frontend expectations (if any),
	// we'll just encode the data directly if it's not nil.
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
