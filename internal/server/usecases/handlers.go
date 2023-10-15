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
		var number string
		var status string
		var amount float32

		if order.Status == "INVALID" || order.Status == "PROCESSED" {
			number = strconv.FormatUint(order.Number, 10)
			status = order.Status
			amount = order.Accrual
		} else {
			accrualResp, err := c.accrual.GetAccrual(r.Context(), order.Number)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to get data from accrual system"))
				return
			}
			if accrualResp == nil {
				accrualResp = &entities.AccrualResponse{
					OrderNumber: strconv.FormatUint(order.Number, 10),
					Status:      order.Status,
					Amount:      0.0,
				}
			} else if accrualResp.Status == "INVALID" || accrualResp.Status == "PROCESSED" {
				accr := entities.Accrual{
					UserID:  *userID,
					OrderID: order.ID,
					Amount:  accrualResp.Amount,
				}
				if err := c.stor.CreateAccrual(r.Context(), &accr); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Failed to store accrual response in the db"))
					return
				}
			}
			number = accrualResp.OrderNumber
			status = accrualResp.Status
			amount = accrualResp.Amount
		}
		val := map[string]any{
			"number":      number,
			"status":      status,
			"accrual":     amount,
			"uploaded_at": order.UploadedAt.Format(time.RFC3339),
		}
		result = append(result, val)
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

func (c *BaseController) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(w, r)
	if userID == nil {
		return
	}
	result, err := c.stor.GetBalance(r.Context(), *userID)
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
	balance, err := c.stor.GetBalance(r.Context(), *userID)
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
	if _, err := c.stor.CreateWithdrawal(r.Context(), withdrawal); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create withdrawal"))
		return
	}
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
