package entities

type ICryptoProvider interface {
	HashPassword(password string) ([]byte, error)
	CheckPassword(password string, hash []byte) error
}
