package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
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
	DB *sql.DB
}

func (m SessionModel) InsertSession(s *model.Sessions) error {
	stmt := table.Sessions.INSERT(table.Sessions.MutableColumns).
		MODEL(s).
		RETURNING(table.Sessions.ID)

	var id []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, &id)
	if err != nil {
		return err
	}

	s.ID = id[0]

	return nil
}

func (m SessionModel) UpdateLastSeen(sessionID int) error {
	stmt := table.Sessions.UPDATE(table.Sessions.LastSeen).
		SET(time.Now().UTC()).
		WHERE(table.Sessions.ID.EQ(helpers.PostgresInt(sessionID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) RemoveSessionByID(sessionID int) error {
	stmt := table.Sessions.DELETE().
		WHERE(table.Sessions.ID.EQ(helpers.PostgresInt(sessionID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) RemoveAllSessionsByUserID(userID int) error {
	stmt := table.Sessions.DELETE().
		WHERE(table.Sessions.UserID.EQ(helpers.PostgresInt(userID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) RemoveAllSessionsByUserIDExceptOne(userID, sessionID int) error {
	stmt := table.Sessions.DELETE().
		WHERE(table.Sessions.UserID.EQ(helpers.PostgresInt(userID)).
			AND(table.Sessions.ID.NOT_EQ(helpers.PostgresInt(sessionID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) GetSessionsByUserID(userID int) ([]*model.Sessions, error) {
	query := postgres.SELECT(table.Sessions.AllColumns).
		FROM(table.Sessions).
		WHERE(table.Sessions.UserID.EQ(helpers.PostgresInt(userID)).
			AND(table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC()))))

	var sessions []*model.Sessions

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &sessions)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (m SessionModel) GetSessionByID(sessionID int) (*model.Sessions, error) {
	query := postgres.SELECT(table.Sessions.AllColumns).
		FROM(table.Sessions).
		WHERE(table.Sessions.ID.EQ(helpers.PostgresInt(sessionID)).
			AND(table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC()))))

	var session model.Sessions

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &session)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchSession
		default:
			return nil, err
		}
	}

	return &session, nil
}
