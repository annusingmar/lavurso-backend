package data

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func ComparePassword(hash []byte, plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, []byte(plaintext))
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
