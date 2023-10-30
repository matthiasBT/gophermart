package entities

import (
	"database/sql"
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

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		UploadedAt string `json:"uploaded_at"`
		*Alias
	}{
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
		Alias:      (*Alias)(o),
	})
}

type Accrual struct {
	ID          int          `db:"id"`
	UserID      int          `db:"user_id"`
	OrderNumber string       `db:"order_number" json:"order"`
	Amount      float32      `db:"amount" json:"sum"`
	ProcessedAt sql.NullTime `db:"processed_at" json:"processed_at"`
}

func (a *Accrual) MarshalJSON() ([]byte, error) {
	var processedAt *string
	if a.ProcessedAt.Valid {
		formatted := a.ProcessedAt.Time.Format(time.RFC3339)
		processedAt = &formatted
	}

	type Alias Accrual
	data := &struct {
		ProcessedAt *string `json:"processed_at"`
		*Alias
	}{
		ProcessedAt: processedAt,
		Alias:       (*Alias)(a),
	}
	return json.Marshal(data)
}

func (a *Accrual) UnmarshalJSON(data []byte) error {
	req := &WithdrawalRequest{}
	if err := json.Unmarshal(data, req); err != nil {
		return err
	}
	a.OrderNumber = req.Number
	a.Amount = req.Sum
	return nil
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
