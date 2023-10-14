package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

func Middleware(logger logging.ILogger, storage entities.Storage) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		checkAuthFn := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && (r.URL.Path == "/api/user/register" || r.URL.Path == "/api/user/login") {
				logger.Infoln("No auth check necessary")
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie("session_token")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Missing session cookie"))
				return
			}
			session, err := storage.FindSession(r.Context(), cookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to find a session"))
				return
			}
			if session == nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Session not found"))
				return
			}
			if time.Now().After(session.ExpiresAt) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Session has expired"))
				return
			}
			logger.Infof("Session is valid, proceeding...")
			ctx := context.WithValue(r.Context(), "user_id", session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(checkAuthFn)
	}
}
