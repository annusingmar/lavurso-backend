package types

import (
	"database/sql/driver"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Hashed    []byte
	Plaintext string
}

func (p Password) Validate(check string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hashed, []byte(check))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (p *Password) Scan(src any) error {
	p.Hashed = src.([]byte)
	return nil
}

func (p Password) Value() (driver.Value, error) {
	return p.Hashed, nil
}
