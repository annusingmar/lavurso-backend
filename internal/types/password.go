package types

type Password struct {
	Hashed    []byte
	Plaintext string
}
