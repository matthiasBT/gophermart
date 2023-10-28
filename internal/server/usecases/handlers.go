package usecases

import (
	"encoding/json"
	"errors"
	"net/http"
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
	tx, err := c.stor.Tx(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create user"))
		return
	}
	defer tx.Commit()
	user, err := c.userRepo.CreateUser(r.Context(), tx, userReq.Login, pwdhash)
	if err != nil {
		defer tx.Rollback()
		if errors.Is(err, entities.ErrLoginAlreadyTaken) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Login is already taken"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to create a new user"))
		}
		return
	}
	session, err := c.userRepo.CreateSession(r.Context(), tx, user, token)
	if err != nil {
		defer tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create a session"))
		return
	}
	authorize(w, session)
}

func (c *BaseController) signIn(w http.ResponseWriter, r *http.Request) {
	userReq := validateUserAuthReq(w, r)
	if userReq == nil {
		return
	}
	user, err := c.userRepo.FindUser(r.Context(), userReq)
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
	tx, err := c.stor.Tx(r.Context())
	defer tx.Commit()
	if err != nil {
		defer tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create a user session"))
		return
	}
	session, err := c.userRepo.CreateSession(r.Context(), tx, user, token)
	if err != nil {
		defer tx.Rollback()
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
	order, existed, err := c.orderRepo.CreateOrder(r.Context(), *userID, *number)
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
	orders, err := c.orderRepo.FindUserOrders(r.Context(), *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to find user's orders"))
		return
	}
	if orders == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	response, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal the result"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (c *BaseController) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(w, r)
	if userID == nil {
		return
	}
	result, err := c.accrualRepo.GetBalance(r.Context(), *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read user balance"))
		return
	}
	response, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal the result"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (c *BaseController) withdraw(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(w, r)
	if userID == nil {
		return
	}
	withdrawal := validateWithdrawal(w, r, *userID)
	if withdrawal == nil {
		return
	}
	balance, err := c.accrualRepo.GetBalance(r.Context(), *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to check user balance"))
		return
	}
	if balance.Current < withdrawal.Amount {
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write([]byte("insufficient funds"))
		return
	}
	if _, err := c.accrualRepo.CreateWithdrawal(r.Context(), withdrawal); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create withdrawal"))
		return
	}
}

func (c *BaseController) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(w, r)
	if userID == nil {
		return
	}
	withdrawals, err := c.accrualRepo.FindUserWithdrawals(r.Context(), *userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to find user's withdrawals"))
		return
	}
	if withdrawals == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	response, err := json.Marshal(withdrawals)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal the result"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
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

func isFinalStatus(status string) bool {
	return status == "INVALID" || status == "PROCESSED"
}
