package entities

import (
	"encoding/json"
	"time"
)

type ContextKey struct {
	Key string
}

type UserAuthRequest struct {
	Login    string
	Password string
}

type User struct {
	ID           int    `db:"id"`
	Login        string `db:"login"`
	PasswordHash []byte `db:"password_hash"`
}

type Session struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

type Order struct {
	ID         int       `db:"id"`
	UserID     int       `db:"user_id"`
	Number     string    `db:"number" json:"number"`
	Status     string    `db:"status" json:"status"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
	Accrual    float32   `db:"accrual" json:"accrual"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		UploadedAt string `json:"uploaded_at"`
		*Alias
	}{
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
		Alias:      (*Alias)(&o),
	})
}

type Accrual struct {
	ID      int     `db:"id"`
	UserID  int     `db:"user_id"`
	OrderID int     `db:"order_id"`
	Amount  float32 `db:"amount"`
}

type AccrualResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Amount      float32 `json:"accrual"`
}

type Balance struct {
	Current   float32 `db:"current" json:"current"`
	WithDrawn float32 `db:"withdrawn" json:"withdrawn"`
}

type WithdrawalRequest struct {
	Number string  `json:"order"`
	Sum    float32 `json:"sum"`
}

type Withdrawal struct {
	ID          int       `db:"id"`
	UserID      int       `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Amount      float32   `db:"amount"`
	ProcessedAt time.Time `db:"processed_at"`
}

func (w *Withdrawal) UnmarshalJSON(data []byte) error {
	req := &WithdrawalRequest{}
	if err := json.Unmarshal(data, req); err != nil {
		return err
	}
	w.OrderNumber = req.Number
	w.Amount = req.Sum
	return nil
}
