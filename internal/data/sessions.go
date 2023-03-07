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

type Session = model.Sessions

type SessionModel struct {
	DB *sql.DB
}

func (m SessionModel) InsertSession(s *Session) error {
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

func (m SessionModel) ExtendSession(sessionID int) error {
	current := time.Now().UTC()

	stmt := table.Sessions.UPDATE(table.Sessions.LastSeen, table.Sessions.Expires).
		SET(current, current.Add(3*time.Minute)).
		WHERE(table.Sessions.ID.EQ(helpers.PostgresInt(sessionID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) ExpireSessionByID(sessionID int) error {
	stmt := table.Sessions.UPDATE(table.Sessions.Expires).
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

func (m SessionModel) ExpireAllSessionsByUserID(userID int) error {
	stmt := table.Sessions.UPDATE(table.Sessions.Expires).
		SET(time.Now().UTC()).
		WHERE(table.Sessions.UserID.EQ(helpers.PostgresInt(userID)).
			AND(table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC()))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) ExpireAllSessionsByUserIDExceptOne(userID, sessionID int) error {
	stmt := table.Sessions.UPDATE(table.Sessions.Expires).
		SET(time.Now().UTC()).
		WHERE(postgres.AND(
			table.Sessions.UserID.EQ(helpers.PostgresInt(userID)),
			table.Sessions.ID.NOT_EQ(helpers.PostgresInt(sessionID)),
			table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC())),
		))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SessionModel) GetSessionsByUserID(userID int) ([]*Session, error) {
	query := postgres.SELECT(table.Sessions.AllColumns).
		FROM(table.Sessions).
		WHERE(table.Sessions.UserID.EQ(helpers.PostgresInt(userID))).
		ORDER_BY(table.Sessions.Expires.DESC())

	var sessions []*Session

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &sessions)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (m SessionModel) GetSessionByID(sessionID int) (*Session, error) {
	query := postgres.SELECT(table.Sessions.AllColumns).
		FROM(table.Sessions).
		WHERE(table.Sessions.ID.EQ(helpers.PostgresInt(sessionID)).
			AND(table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC()))))

	var session Session

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
