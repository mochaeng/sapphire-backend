package cryptoutils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
)

func GenerateRandomString(numBytes int) (string, error) {
	bytes := make([]byte, numBytes)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}

func GetSessionID(token string) string {
	hashedToken := sha256.Sum256([]byte(token))
	sessionID := hex.EncodeToString(hashedToken[:])
	return sessionID
}
