package source

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	IncorrectDirectUserStruct = errors.New("incorrect direction in user data")
	incorrectHashStatus       = errors.New("incorrect hash status")
)

const (
	UserCreate = iota + 1
	UserConnect

	NewLogin
	NewPassword
	NewName
	NewEmail

	UserDelete
)

type ChangeLogin struct {
	Login string `json:"login"`
}

const (
	Hashed   = 100
	NoHashed = 110
)

type ChangePassword struct {
	Hashed      int    `json:"hashed"`
	PasswordOne string `json:"password_one"`
	PasswordTwo string `json:"password_two"`
}

func (c *ChangePassword) HashPassword() error {
	if c.Hashed == Hashed {
		return nil
	}

	if c.Hashed == NoHashed {
		c.PasswordOne = HashData(c.PasswordOne)
		c.PasswordTwo = c.PasswordOne
		c.Hashed = Hashed

		return nil
	}

	return incorrectHashStatus
}

func HashData(line string) string {
	hash := sha256.Sum256([]byte(line))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}

type ChangeName struct {
	Name    string `json:"first_name"`
	Surname string `json:"last_name,omitempty"`
}

type ChangeEmail struct {
	Email string `json:"email"`
}

type UserSourceData struct {
	Direct         int `json:"direct"`
	ID             int `json:"id,omitempty"`
	ChangeLogin    `json:"change_login,omitempty "`
	ChangePassword `json:"change_password,omitempty"`
	ChangeName     `json:"change_name,omitempty"`
	ChangeEmail    `json:"change_email,omitempty"`
}

type User struct {
	ID           int    `json:"id,omitempty" db:"id"`
	Login        string `json:"login" db:"login"`
	HashPassword string `json:"-" db:"hashed_password"`
	Name         string `json:"name" db:"name"`
	Surname      string `json:"surname,omitempty" db:"surname,omitempty"`
	Email        string `json:"email" db:"email"`
}

type Message struct {
	Msg string `json:"message"`
}
