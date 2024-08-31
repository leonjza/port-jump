package secrets

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
)

func GenerateTOTPSecret(length int) (string, error) {
	// Base32 encoding requires 5 bits per character.
	// To generate a string of `length` characters, we need 5 * length bits.
	// So, 5 * length / 8 bytes.
	byteLength := (length * 5) / 8

	randomBytes := make([]byte, byteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Encode the random bytes to a Base32 string
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Trim the string to the desired length
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}
