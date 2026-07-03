// internal/middleware/metrics.go
package middleware

import (
	"net/http"
	"strconv"
	"task_tracker/pkg/metrics"
	"time"
)

// responseWriter — враппер чтобы перехватить статус код
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.status)
		path := r.Pattern // chi даёт шаблон /tasks/{id}, а не /tasks/123

		metrics.RequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		metrics.RequestDuration.WithLabelValues(r.Method, path).Observe(duration)

		if rw.status >= 400 {
			metrics.ErrorsTotal.WithLabelValues(r.Method, path, status).Inc()
		}
	})
}
