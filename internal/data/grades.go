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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoSuchGrade             = errors.New("no such grade")
	ErrIdentifierAlreadyExists = errors.New("identifier already exists")
)

type Grade struct {
	ID         *int    `json:"id,omitempty"`
	Identifier *string `json:"identifier,omitempty"`
	Value      *int    `json:"value,omitempty"`
}

type GradeModel struct {
	DB *sql.DB
}

func (m GradeModel) AllGrades() ([]*model.Grades, error) {
	query := postgres.SELECT(table.Grades.AllColumns).
		FROM(table.Grades).
		ORDER_BY(table.Grades.ID.ASC())

	var grades []*model.Grades

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &grades)
	if err != nil {
		return nil, err
	}

	return grades, nil
}

func (m GradeModel) GetGradeByID(gradeID int) (*model.Grades, error) {
	query := postgres.SELECT(table.Grades.AllColumns).
		FROM(table.Grades).
		WHERE(table.Grades.ID.EQ(helpers.PostgresInt(gradeID)))

	var grade model.Grades

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &grade)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchGrade
		default:
			return nil, err
		}
	}

	return &grade, nil
}

func (m GradeModel) UpdateGrade(g *model.Grades) error {
	stmt := table.Grades.UPDATE(table.Grades.Identifier, table.Grades.Value).
		MODEL(g).
		WHERE(table.Grades.ID.EQ(helpers.PostgresInt(g.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrIdentifierAlreadyExists
		} else {
			return err
		}
	}

	return nil
}

func (m GradeModel) InsertGrade(g *Grade) error {
	stmt := table.Grades.INSERT(table.Grades.Identifier, table.Grades.Value).
		MODEL(g).
		RETURNING(table.Grades.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, g)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrIdentifierAlreadyExists
		} else {
			return err
		}
	}

	return nil
}
