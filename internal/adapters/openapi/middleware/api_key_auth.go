package middleware

import (
	"net/http"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/dariomba/screen-go/internal/logger"
)

func APIKeyAuthMiddleware(apiKeys []string) func(http.Handler) http.Handler {
	if len(apiKeys) == 0 {
		logger.Warn().Msg("no API key is provided, the API will be accessible without authentication")
	}
	keySet := make(map[string]struct{}, len(apiKeys))
	for _, k := range apiKeys {
		if k != "" {
			keySet[k] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(keySet) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := keySet[r.Header.Get("X-API-Key")]; !ok {
				logger.Ctx(r.Context()).Error().
					Str("error_type", "authentication_error").
					Msg("unauthorized access attempt with invalid API key")
				openapi.WriteErrorJSON(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
