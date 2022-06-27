package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoSuchJournal = errors.New("no such journal")
)

type Journal struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	TeacherID int    `json:"teacher_id"`
	SubjectID int    `json:"subject_id"`
	Archived  bool   `json:"archived"`
}

type JournalModel struct {
	DB *sql.DB
}

func (m JournalModel) AllJournals() ([]*Journal, error) {
	query := `SELECT id, name, teacher_id, subject_id, archived
	FROM journals
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var journals []*Journal

	for rows.Next() {
		var journal Journal
		err = rows.Scan(
			&journal.ID,
			&journal.Name,
			&journal.TeacherID,
			&journal.SubjectID,
			&journal.Archived,
		)
		if err != nil {
			return nil, err
		}

		journals = append(journals, &journal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return journals, nil
}

func (m JournalModel) GetJournalByID(journalID int) (*Journal, error) {
	query := `SELECT id, name, teacher_id, subject_id, archived
	FROM journals
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var journal Journal

	err := m.DB.QueryRowContext(ctx, query, journalID).Scan(
		&journal.ID,
		&journal.Name,
		&journal.TeacherID,
		&journal.SubjectID,
		&journal.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchJournal
		default:
			return nil, err
		}
	}

	return &journal, nil
}

func (m JournalModel) InsertJournal(j *Journal) error {
	stmt := `INSERT INTO journals
	(name, teacher_id, subject_id, archived)
	VALUES
	($1, $2, $3, $4)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, j.Name, j.TeacherID, j.SubjectID, j.Archived).Scan(&j.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m JournalModel) UpdateJournal(j *Journal) error {
	stmt := `UPDATE journals
	SET (name, teacher_id, subject_id, archived)
	= ($1, $2, $3, $4)
	WHERE id = $5`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, j.Name, j.TeacherID, j.SubjectID, j.Archived, j.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m JournalModel) GetJournalsForTeacher(teacherID int) ([]*Journal, error) {
	query := `SELECT id, name, teacher_id, subject_id, archived
	FROM journals
	WHERE teacher_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var journals []*Journal

	for rows.Next() {
		var journal Journal
		err = rows.Scan(
			&journal.ID,
			&journal.Name,
			&journal.TeacherID,
			&journal.SubjectID,
			&journal.Archived,
		)
		if err != nil {
			return nil, err
		}

		journals = append(journals, &journal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return journals, nil
}
