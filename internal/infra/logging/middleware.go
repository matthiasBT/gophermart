package logging

import (
	"net/http"
	"time"
)

type responseMetadata struct {
	status int
	size   int
}

type extendedWriter struct {
	http.ResponseWriter
	response *responseMetadata
}

func (w *extendedWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.response.size += size
	return size, err
}

func (w *extendedWriter) WriteHeader(status int) {
	w.response.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Middleware(logger ILogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			extWriter := &extendedWriter{
				ResponseWriter: w,
				response: &responseMetadata{
					status: http.StatusOK,
					size:   0,
				},
			}
			next.ServeHTTP(extWriter, r)
			duration := time.Since(start)
			logger.Infof("Served: %s %s, %v\n", r.Method, r.RequestURI, duration)
			logger.Infof(
				"Response: [%d] %d bytes \n", extWriter.response.status, extWriter.response.size,
			)
		}
		return http.HandlerFunc(logFn)
	}
}
