package usecases

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/matthiasBT/gophermart/internal/infra/config"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

func (c *BaseController) register(w http.ResponseWriter, r *http.Request) {
	userReq := validateUserAuthReq(w, r)
	if userReq == nil {
		return
	}
	pwdhash, err := c.crypto.HashPassword(userReq.Password)
	if err != nil {
		return
	}
	token := generateSessionToken()
	_, session, err := c.stor.CreateUser(r.Context(), userReq.Login, pwdhash, token)
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

func (c *BaseController) signIn(w http.ResponseWriter, r *http.Request) {
	userReq := validateUserAuthReq(w, r)
	if userReq == nil {
		return
	}
	user, err := c.stor.FindUser(r.Context(), userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to find the user"))
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("User doesn't exist"))
		return
	}
	if err := c.crypto.CheckPassword(userReq.Password, user.PasswordHash); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect password"))
		return
	}
	token := generateSessionToken()
	session, err := c.stor.CreateSession(r.Context(), nil, user, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create a user session"))
		return
	}
	authorize(w, session)
}

func (c *BaseController) createOrder(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(w, r)
	if userID == nil {
		return
	}
	number := validateOrderNumber(w, r)
	if number == nil {
		return
	}
	order, existed, err := c.stor.CreateOrder(r.Context(), *userID, *number)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create an order"))
		return
	}
	if existed {
		if order.UserID == *userID {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("already created"))
		} else {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("already created by another user"))
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (c *BaseController) getOrders(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(w, r)
	if userID == nil {
		return
	}
	orders, err := c.stor.FindUserOrders(r.Context(), *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to find user's orders"))
		return
	}
	if orders == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var result []map[string]any
	for _, order := range orders {
		val := map[string]any{
			"number":      strconv.FormatUint(order.Number, 10),
			"status":      order.Status,
			"accrual":     0,
			"uploaded_at": order.UploadedAt.Format(time.RFC3339),
		}
		result = append(result, val)
	}
	response, err := json.Marshal(result)
	if err != nil {
		c.logger.Errorf("Failed to marshal the response: %s", err.Error())
		return
	}
	w.Write(response)
}

func authorize(w http.ResponseWriter, session *entities.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		Expires:  time.Now().Add(config.SessionTTL),
		HttpOnly: true,  // Protect against XSS attacks
		Secure:   false, // Should be true in production to send only over HTTPS
	})
}

func getUserID(w http.ResponseWriter, r *http.Request) *int {
	userID := r.Context().Value(entities.ContextKey{Key: "user_id"})
	if userID == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to find the user_id in the context"))
		return nil
	}
	res := userID.(int)
	return &res
}
