package entities

type ICryptoProvider interface {
	HashPassword(password string) ([]byte, error)
}
