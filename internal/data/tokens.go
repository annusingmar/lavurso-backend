package data

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	TokenAuthentication = "auth"
)

type Token struct {
	ID        int
	Hash      []byte
	Plaintext string
	UserID    int
	Type      string
	Expires   time.Time
}

type TokenModel struct {
	DB *pgxpool.Pool
}

func generateNewToken(userID int, expiresIn time.Duration, tokenType string) (*Token, error) {
	randomData := make([]byte, 16)

	_, err := rand.Read(randomData)
	if err != nil {
		return nil, err
	}

	token := &Token{
		Plaintext: base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomData),
		UserID:    userID,
		Expires:   time.Now().UTC().Add(expiresIn),
		Type:      tokenType,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))

	token.Hash = hash[:]

	return token, nil
}

func (m TokenModel) InsertToken(token *Token) error {
	stmt := `INSERT INTO tokens
	(hash, user_id, type, expires)
	VALUES
	($1, $2, $3, $4)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, token.Hash, token.UserID, token.Type, token.Expires).Scan(&token.ID)
	if err != nil {
		return err
	}

	return nil
}
