package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRawKey() (string, error) {
	buf := make([]byte, 24)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return "sm_" + hex.EncodeToString(buf), nil
}
