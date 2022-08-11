package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrNoSuchSession = errors.New("no such session")
)

type Session struct {
	ID             int       `json:"id"`
	TokenHash      []byte    `json:"-"`
	TokenPlaintext string    `json:"token,omitempty"`
	UserID         int       `json:"user_id"`
	Expires        time.Time `json:"expires"`
	LoginIP        string    `json:"login_ip"`
	LoginBrowser   string    `json:"login_browser"`
	LoggedIn       time.Time `json:"logged_in"`
	LastSeen       time.Time `json:"last_seen"`
}

type SessionModel struct {
	DB *pgxpool.Pool
}

func (s *Session) AddNewTokenToSession() error {
	randomData := make([]byte, 16)

	_, err := rand.Read(randomData)
	if err != nil {
		return err
	}

	s.TokenPlaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomData)
	hash := sha256.Sum256([]byte(s.TokenPlaintext))
	s.TokenHash = hash[:]

	return nil
}

func (m SessionModel) InsertSession(session *Session) error {
	stmt := `INSERT INTO sessions
	(token_hash, user_id, expires, login_ip, login_browser, logged_in, last_seen)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, session.TokenHash, session.UserID, session.Expires, session.LoginIP, session.LoginBrowser, session.LoggedIn, session.LastSeen).Scan(&session.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) UpdateLastSeen(sessionID int) error {
	stmt := `UPDATE sessions
	SET last_seen = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, time.Now().UTC(), sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) RemoveSessionByID(sessionID int) error {
	stmt := `DELETE FROM sessions
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) RemoveAllSessionsByUserID(userID int) error {
	stmt := `DELETE FROM sessions
	WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) GetSessionsByUserID(userID int) ([]*Session, error) {
	query := `SELECT
	id, token_hash, user_id, expires, login_ip, login_browser, logged_in, last_seen
	FROM sessions
	WHERE user_id = $1
	AND expires > $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sessions []*Session

	for rows.Next() {
		var session Session

		err = rows.Scan(
			&session.ID,
			&session.TokenHash,
			&session.UserID,
			&session.Expires,
			&session.LoginIP,
			&session.LoginBrowser,
			&session.LoggedIn,
			&session.LastSeen,
		)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, &session)
	}

	if rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (m SessionModel) GetSessionByID(sessionID int) (*Session, error) {
	query := `SELECT
	id, token_hash, user_id, expires, login_ip, login_browser, logged_in, last_seen
	FROM sessions
	WHERE id = $1
	AND expires > $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var session Session

	err := m.DB.QueryRow(ctx, query, sessionID, time.Now().UTC()).Scan(
		&session.ID,
		&session.TokenHash,
		&session.UserID,
		&session.Expires,
		&session.LoginIP,
		&session.LoginBrowser,
		&session.LoggedIn,
		&session.LastSeen,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
