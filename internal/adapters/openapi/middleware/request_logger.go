package middleware

import (
	"net/http"
	"time"

	"github.com/dariomba/screen-go/internal/logger"
	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// RequestLogger logs HTTP requests with structured logging and adds request ID to context
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.New().String()

		ctx := logger.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		w.Header().Set("X-Request-ID", requestID)

		// Log incoming request
		logger.Ctx(ctx).Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("Incoming request")

		next.ServeHTTP(wrapped, r)

		// Log completed request
		duration := time.Since(start)

		logger.Ctx(ctx).Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.status).
			Int("size_bytes", wrapped.size).
			Dur("duration", duration).
			Msg("Request completed")
	})
}

// Recovery logs panics and returns 500
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Ctx(r.Context()).Error().
					Interface("panic", err).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("Panic recovered")

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal server error"}`)) //nolint:errcheck
			}
		}()
		next.ServeHTTP(w, r)
	})
}
