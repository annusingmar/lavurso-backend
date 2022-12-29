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
	ErrNoSuchSubject = errors.New("no such subject")
)

type SubjectModel struct {
	DB *sql.DB
}

func (m SubjectModel) AllSubjects() ([]*model.Subjects, error) {
	query := postgres.SELECT(table.Subjects.AllColumns).
		FROM(table.Subjects).
		ORDER_BY(table.Subjects.ID.ASC())

	var subjects []*model.Subjects

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &subjects)
	if err != nil {
		return nil, err
	}

	return subjects, nil
}

func (m SubjectModel) InsertSubject(s *model.Subjects) error {
	stmt := table.Subjects.INSERT(table.Subjects.MutableColumns).
		MODEL(s).
		RETURNING(table.Subjects.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, s)
	if err != nil {
		return err
	}

	return nil
}

func (m SubjectModel) UpdateSubject(s *model.Subjects) error {
	stmt := table.Subjects.UPDATE(table.Subjects.Name).
		MODEL(s).
		WHERE(table.Subjects.ID.EQ(helpers.PostgresInt(s.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SubjectModel) GetSubjectByID(subjectID int) (*model.Subjects, error) {
	query := postgres.SELECT(table.Subjects.AllColumns).
		FROM(table.Subjects).
		WHERE(table.Subjects.ID.EQ(helpers.PostgresInt(subjectID)))

	var subject model.Subjects

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &subject)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchSubject
		default:
			return nil, err
		}
	}

	return &subject, nil
}
