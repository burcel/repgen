package security

import (
	"crypto/rand"
	"encoding/hex"
)

// Return string representation of randomized byte array with respect to given length
// 1 byte consists of 2 length of string e.g. 00, 01, FF etc.
func GenerateRandomHex(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
