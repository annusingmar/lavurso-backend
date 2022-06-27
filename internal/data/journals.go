package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoSuchJournal        = errors.New("no such journal")
	ErrUserAlreadyInJournal = errors.New("user is already part of journal")
	ErrJournalArchived      = errors.New("journal is archived")
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

func (m JournalModel) InsertUserIntoJournal(userID, journalID int) error {
	stmt := `INSERT INTO
	users_journals
	(user_id, journal_id)
	VALUES
	($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, userID, journalID)
	if err != nil {
		switch {
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_journals_pkey" (SQLSTATE 23505)`:
			return ErrUserAlreadyInJournal
		default:
			return err
		}
	}

	return nil
}

func (m JournalModel) GetUsersByJournalID(journalID int) ([]*User, error) {
	query := `SELECT id, name, email, password, role, created_at, active, version
	FROM users u
	INNER JOIN users_journals uj
	ON uj.user_id = u.id
	WHERE uj.journal_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password.Hashed,
			&user.Role,
			&user.CreatedAt,
			&user.Active,
			&user.Version,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m JournalModel) GetJournalsByUserID(userID int) ([]*Journal, error) {
	query := `SELECT id, name, teacher_id, subject_id, archived
	FROM journals j
	INNER JOIN users_journals uj
	ON uj.journal_id = j.id
	WHERE uj.user_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
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
