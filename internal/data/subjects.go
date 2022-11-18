package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoSuchSubject = errors.New("no such subject")
)

type Subject struct {
	ID   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type SubjectModel struct {
	DB *sql.DB
}

func (m SubjectModel) AllSubjects() ([]*Subject, error) {
	query := `SELECT id, name
	FROM subjects
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var subjects []*Subject

	for rows.Next() {
		var subject Subject
		err = rows.Scan(
			&subject.ID,
			&subject.Name,
		)
		if err != nil {
			return nil, err
		}

		subjects = append(subjects, &subject)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subjects, nil
}

func (m SubjectModel) InsertSubject(s *Subject) error {
	stmt := `INSERT INTO subjects
	(name) VALUES ($1)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, s.Name).Scan(&s.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m SubjectModel) UpdateSubject(s *Subject) error {
	stmt := `UPDATE subjects
	SET name
	= $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, s.Name, s.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m SubjectModel) GetSubjectByID(subjectID int) (*Subject, error) {
	query := `SELECT id, name
	FROM subjects
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var subject Subject

	err := m.DB.QueryRowContext(ctx, query, subjectID).Scan(
		&subject.ID,
		&subject.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchSubject
		default:
			return nil, err
		}
	}

	return &subject, nil
}
