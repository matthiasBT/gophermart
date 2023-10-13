package usecases

import (
	"errors"
	"net/http"

	"github.com/matthiasBT/gophermart/internal/server/entities"
)

func (c *BaseController) register(w http.ResponseWriter, r *http.Request) {
	userReq := validateUser(w, r)
	if userReq == nil {
		return
	}
	user, err := c.Stor.CreateUser(r.Context(), userReq)
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
	// todo: set cookie
	c.Logger.Infof("User: %v", user)
}
