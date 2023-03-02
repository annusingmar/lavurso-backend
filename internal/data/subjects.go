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
	ErrNoSuchSubject     = errors.New("no such subject")
	ErrCantDeleteSubject = errors.New("can't delete subject with journals")
)

type Subject = model.Subjects

type SubjectExt struct {
	Subject
	JournalCount *int `json:"journal_count,omitempty"`
}

type SubjectWithMarks struct {
	Subject
	Marks map[int][]*MarkExt `json:"marks,omitempty"`
}

type SubjectModel struct {
	DB *sql.DB
}

func (m SubjectModel) AllSubjects() ([]*SubjectExt, error) {
	query := postgres.SELECT(table.Subjects.AllColumns, postgres.COUNT(table.Journals.ID).AS("subjectext.journal_count")).
		FROM(table.Subjects.
			LEFT_JOIN(table.Journals, table.Journals.SubjectID.EQ(table.Subjects.ID))).
		GROUP_BY(table.Subjects.ID).
		ORDER_BY(table.Subjects.ID.ASC())

	var subjects []*SubjectExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &subjects)
	if err != nil {
		return nil, err
	}

	return subjects, nil
}

func (m SubjectModel) InsertSubject(s *Subject) error {
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

func (m SubjectModel) UpdateSubject(s *SubjectExt) error {
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

func (m SubjectModel) DeleteSubject(subjectID int) error {
	stmt := table.Subjects.DELETE().
		WHERE(table.Subjects.ID.EQ(helpers.PostgresInt(subjectID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m SubjectModel) GetSubjectByID(subjectID int, getJournalCount bool) (*SubjectExt, error) {
	var query postgres.SelectStatement

	if getJournalCount {
		query = postgres.SELECT(table.Subjects.AllColumns, postgres.COUNT(table.Journals.ID).AS("subjectext.journal_count")).
			FROM(table.Subjects.
				LEFT_JOIN(table.Journals, table.Journals.SubjectID.EQ(table.Subjects.ID))).
			GROUP_BY(table.Subjects.ID)
	} else {
		query = postgres.SELECT(table.Subjects.AllColumns).
			FROM(table.Subjects)
	}

	query = query.WHERE(table.Subjects.ID.EQ(helpers.PostgresInt(subjectID)))

	var subject SubjectExt

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
