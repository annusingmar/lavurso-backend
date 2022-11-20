package types

import "database/sql/driver"

type Password struct {
	Hashed    []byte
	Plaintext string
}

func (p *Password) Scan(src any) error {
	p.Hashed = src.([]byte)
	return nil
}

func (p Password) Value() (driver.Value, error) {
	return p.Hashed, nil
}
