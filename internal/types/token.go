package types

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base32"
)

type Token struct {
	Hashed    []byte `json:"-"`
	Plaintext string `json:"value"`
}

func (t *Token) NewToken() error {
	randomData := make([]byte, 16)

	_, err := rand.Read(randomData)
	if err != nil {
		return err
	}

	t.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomData)
	hash := sha256.Sum256([]byte(t.Plaintext))
	t.Hashed = hash[:]

	return nil
}

func (t *Token) Scan(src any) error {
	t.Hashed = src.([]byte)
	return nil
}

func (t Token) Value() (driver.Value, error) {
	return t.Hashed, nil
}
