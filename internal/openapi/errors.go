package openapi

import (
	"encoding/json"
	"log"
	"net/http"
)

type HTTPErrorResponse struct {
	Error string `json:"error"`
}

func WriteErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := HTTPErrorResponse{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// If JSON encoding fails, log it but don't try to write again
		// (headers already sent)
		log.Printf("Failed to encode error response: %v", err)
	}
}
