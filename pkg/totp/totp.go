package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"strings"
	"time"
)

// ref: https://datatracker.ietf.org/doc/html/rfc6238
type Totp struct {
	secret   string
	interval int64
}

// NewTotp creates a new Totp struct
func NewTotp(secret string, interval int64) (*Totp, error) {

	if secret == "" {
		return nil, errors.New("secret cannot be empty")
	}

	if interval == 0 {
		return nil, errors.New("interval cannot be zero")
	}

	return &Totp{
		secret:   secret,
		interval: interval,
	}, nil
}

// Code returns an integer of a calculated HMAC
func (t *Totp) Code() (uint32, error) {
	secret := strings.ToUpper(t.secret)
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return 0, err
	}

	counter := time.Now().Unix() / t.interval

	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], uint64(counter))

	hash := hmac.New(sha1.New, key)
	hash.Write(counterBytes[:])
	hmacHash := hash.Sum(nil)

	offset := hmacHash[len(hmacHash)-1] & 0x0F

	truncatedHash := hmacHash[offset : offset+4]
	truncatedHash[0] &= 0x7F // Ensure the most significant bit is 0 (to avoid a negative number)

	return binary.BigEndian.Uint32(truncatedHash), nil
}

// Generate generates a typical 6-digit HOTP
func (t *Totp) Generate() (int, error) {
	code, err := t.Code()
	if err != nil {
		return 0, err
	}

	return int(code % 1000000), nil
}

// GenerateTCPPort generates a HOTP within the TCP high-port range.
func (t *Totp) GenerateTCPPort() (int, error) {
	code, err := t.Generate()
	if err != nil {
		return 0, nil
	}

	minPort := 1024
	maxPort := 65535
	return int(code%(maxPort-minPort+1) + minPort), nil
}
