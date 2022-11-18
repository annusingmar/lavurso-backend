package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotValidAbsence = errors.New("not valid absence for user")
	ErrNoSuchAbsence   = errors.New("no such absence")
	ErrAbsenceExcused  = errors.New("absence already excused")
	ErrNoSuchExcuse    = errors.New("no such excuse")
)

type Excuse struct {
	MarkID *int       `json:"mark_id,omitempty"`
	Excuse *string    `json:"excuse,omitempty"`
	By     *User      `json:"by,omitempty"`
	At     *time.Time `json:"at,omitempty"`
}

type AbsenceModel struct {
	DB *sql.DB
}

func (m AbsenceModel) InsertExcuse(excuse *Excuse) error {
	stmt := `INSERT INTO excuses
	(mark_id, excuse, by, at)
	VALUES
	($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, excuse.MarkID, excuse.Excuse, excuse.By.ID, excuse.At)
	if err != nil {
		return err
	}

	return nil
}

func (m AbsenceModel) DeleteExcuseByMarkID(markID int) error {
	stmt := `DELETE FROM excuses
	WHERE mark_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, markID)
	if err != nil {
		return err
	}

	return nil
}
