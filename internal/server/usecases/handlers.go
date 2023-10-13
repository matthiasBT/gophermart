package usecases

import (
	"crypto/rand"
	"encoding/base64"
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
	_, err := c.Stor.CreateUser(r.Context(), userReq)
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
	authorize(w)
	return
}

func generateSessionToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func authorize(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    generateSessionToken(),
		Path:     "/",
		Expires:  time.Now().Add(120 * time.Second),
		HttpOnly: true,  // Protect against XSS attacks
		Secure:   false, // Should be true in production to send only over HTTPS
	})
}
