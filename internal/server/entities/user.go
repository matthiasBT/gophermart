package entities

type UserCreateRequest struct {
	Login    string
	Password string
}

type User struct {
	ID           int    `db:"id"`
	Login        string `db:"login"`
	PasswordHash string `db:"password_hash"`
}
