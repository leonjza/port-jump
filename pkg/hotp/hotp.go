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

// ref: https://www.ietf.org/rfc/rfc4226.txt
type Hotp struct {
	secret   string
	interval int64
}

// NewTotp creates a new Totp struct
func NewTotp(secret string, interval int64) (*Hotp, error) {
	if secret == "" {
		return nil, errors.New("secret cannot be empty")
	}

	if interval == 0 {
		return nil, errors.New("interval cannot be zero")
	}

	return &Hotp{
		secret:   secret,
		interval: interval,
	}, nil
}

// Code returns an integer of a calculated HMAC
func (h *Hotp) Code() (uint32, error) {
	secret := strings.ToUpper(h.secret)
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return 0, err
	}

	counter := time.Now().Unix() / h.interval

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
func (h *Hotp) Generate() (int, error) {
	code, err := h.Code()
	if err != nil {
		return 0, err
	}

	return int(code % 1000000), nil
}

// GenerateTCPPort generates a HOTP within the TCP high-port range.
func (h *Hotp) GenerateTCPPort() (int, error) {
	code, err := h.Code()
	if err != nil {
		return 0, err
	}

	var minPort uint32 = 1024
	var maxPort uint32 = 65535

	return int(code%(maxPort-minPort+1) + minPort), nil
}
