package cryptoutils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"math/big"
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

// GenerateRandomSuffix generates a string with only numbers and lowercase letters
func GenerateRandomSuffix(length int) (string, error) {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	suffix := make([]byte, length)
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		suffix[i] = chars[num.Int64()]
	}
	return string(suffix), nil
}
