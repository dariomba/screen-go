package openapi

import (
	"encoding/json"
	"net/http"

	"github.com/dariomba/screen-go/internal/logger"
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
		logger.Error().
			Err(err).
			Str("error_type", "json_encoding_error").
			Msg("Failed to encode error response")
	}
}
