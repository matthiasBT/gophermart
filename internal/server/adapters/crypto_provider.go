package adapters

import (
	"github.com/matthiasBT/gophermart/internal/infra/logging"
	"golang.org/x/crypto/bcrypt"
)

type CryptoProvider struct {
	Logger logging.ILogger
}

func (cr *CryptoProvider) HashPassword(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		cr.Logger.Errorf("Failed to hash password: %s", err.Error())
		return nil, err
	}
	return hashedPassword, nil
}
