package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Password      []byte `json:"-"` // - means that this field will not be returned in the response
	Phone         string `json:"phone"`
	Picture       string `json:"picture"`
	PublicAddress string `json:"public_address"`
}

func (u *User) SetPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	u.Password = hashedPassword
}

func (u *User) ComparePassword(password string) error {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	return err
}
