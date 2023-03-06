package types

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"time"
)

type TOTPSecret string

func GenerateSecret() (TOTPSecret, error) {
	b := make([]byte, 20)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return TOTPSecret(s), nil
}

func (secret *TOTPSecret) Validate(otp int) (bool, error) {
	key, err := base32.StdEncoding.DecodeString(string(*secret))
	if err != nil {
		return false, err
	}

	timestamp := time.Now().Unix() / 30
	hm := hmac.New(sha1.New, key)
	binary.Write(hm, binary.BigEndian, timestamp)

	h := hm.Sum(nil)
	offs := h[len(h)-1] & 0xF
	truncatedHash := binary.BigEndian.Uint32(h[offs : offs+4])

	code := (truncatedHash & 0x7FFFFFFF) % 1000000

	return otp == int(code), nil
}
