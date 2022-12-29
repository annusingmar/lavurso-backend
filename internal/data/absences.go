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

type Excuse = model.Excuses

type ExcuseExt struct {
	Excuse
	By *User `json:"by,omitempty" alias:"excuser"`
}

type AbsenceModel struct {
	DB *sql.DB
}

func (m AbsenceModel) InsertExcuse(excuse *Excuse) error {
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
