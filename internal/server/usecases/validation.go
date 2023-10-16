package usecases

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/matthiasBT/gophermart/internal/server/entities"
)

const MinLoginLength = 1
const MinPasswordLength = 1
const MinOrderNumberLength = 1

func validateUserAuthReq(w http.ResponseWriter, r *http.Request) *entities.UserAuthRequest {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Supply data as JSON"))
		return nil
	}
	var userReq entities.UserAuthRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read request body"))
		return nil
	}
	if err := json.Unmarshal(body, &userReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse user create request"))
		return nil
	}
	if len(userReq.Login) < MinLoginLength || len(userReq.Password) < MinPasswordLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Login or password is too short"))
		return nil
	}
	return &userReq
}

func validateOrderNumber(w http.ResponseWriter, r *http.Request) *uint64 {
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Supply data as plaintext"))
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read request body"))
		return nil
	}
	number := string(body)
	if err := validatePlainOrderNumber(w, number); err != nil {
		return nil
	}
	res, _ := strconv.ParseUint(number, 10, 64)
	return &res
}

func validateWithdrawal(w http.ResponseWriter, r *http.Request, userID int) *entities.Withdrawal {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Supply data as JSON"))
		return nil
	}
	var withdrawal entities.Withdrawal
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read request body"))
		return nil
	}
	if err := json.Unmarshal(body, &withdrawal); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to parse withdrawal request"))
		return nil
	}
	number := strconv.FormatUint(withdrawal.OrderNumber, 10)
	if err := validatePlainOrderNumber(w, number); err != nil {
		return nil
	}
	withdrawal.UserID = userID
	return &withdrawal
}

func validatePlainOrderNumber(w http.ResponseWriter, number string) error {
	if len(number) < MinOrderNumberLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The order number is too short"))
		return errors.New("number is too short")
	}
	if err := goluhn.Validate(number); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("Invalid order number: Luhn algorithm check failed"))
		return errors.New("non-Luhn order number")
	}
	return nil
}
