package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// func HashPassword(password string) (string, error): Hash the password using the bcrypt.GenerateFromPassword function. Bcrypt is a secure hash function that is intended for use with passwords.
// func CheckPasswordHash(password, hash string) error: Use the bcrypt.CompareHashAndPassword function to compare the password that the user entered in the HTTP request with the password that is stored in the database.

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
