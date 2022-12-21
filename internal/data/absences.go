package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
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

type NExcuse struct {
	model.Excuses
	By *model.Users `json:"by,omitempty" alias:"excuser"`
}

type AbsenceModel struct {
	DB *sql.DB
}

func (m AbsenceModel) InsertExcuse(excuse *model.Excuses) error {
	stmt := table.Excuses.INSERT(table.Excuses.AllColumns).
		MODEL(excuse)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m AbsenceModel) DeleteExcuseByMarkID(markID int) error {
	stmt := table.Excuses.DELETE().
		WHERE(table.Excuses.MarkID.EQ(helpers.PostgresInt(markID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
