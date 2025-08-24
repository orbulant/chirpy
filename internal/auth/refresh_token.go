package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	buffer := make([]byte, 32)

	_, err := rand.Read(buffer)

	if err != nil {
		return "", err
	}

	encoded := hex.EncodeToString(buffer)

	return encoded, nil
}
