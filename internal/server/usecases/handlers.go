package usecases

import (
	"errors"
	"net/http"
	"time"

	"github.com/matthiasBT/gophermart/internal/server/entities"
)

func (c *BaseController) register(w http.ResponseWriter, r *http.Request) {
	userReq := validateUser(w, r)
	if userReq == nil {
		return
	}
	token := generateSessionToken()
	_, session, err := c.Stor.CreateUser(r.Context(), userReq, token)
	if err != nil {
		if errors.Is(err, entities.ErrLoginAlreadyTaken) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Login is already taken"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to create a new user"))
		}
		return
	}
	authorize(w, session)
}

func authorize(w http.ResponseWriter, session *entities.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		Expires:  time.Now().Add(120 * time.Second), // TODO: check expiration on server
		HttpOnly: true,                              // Protect against XSS attacks
		Secure:   false,                             // Should be true in production to send only over HTTPS
	})
}
