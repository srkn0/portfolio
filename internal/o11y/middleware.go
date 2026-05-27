package o11y

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLog returns chi middleware that logs each request as one structured
// Info record with method, route pattern, status, duration and request ID.
// Use this instead of chi.Logger so logs are structured and queryable in Loki.
func RequestLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sr, r)

			route := RouteFromContext(r)
			if route == "" {
				route = r.URL.Path
			}

			logger.LogAttrs(r.Context(), slog.LevelInfo, "http_request",
				slog.String("event", "http_request"),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("route", route),
				slog.Int("status", sr.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
				slog.String("trace_id", traceIDFrom(r.Context())),
				slog.String("remote_ip", clientIP(r)),
				slog.String("user_agent", r.UserAgent()),
				slog.String("referer", r.Referer()),
			)
		})
	}
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return v
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	return r.RemoteAddr
}
