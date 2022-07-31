package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNoSuchJournal        = errors.New("no such journal")
	ErrUserAlreadyInJournal = errors.New("user is already part of journal")
	ErrJournalArchived      = errors.New("journal is archived")
	ErrUserNotInJournal     = errors.New("user not in journal")
)

type Journal struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// TeacherID int    `json:"teacher_id"`
	// SubjectID int    `json:"subject_id"`
	Teacher  *User    `json:"teacher"`
	Subject  *Subject `json:"subject"`
	Archived bool     `json:"archived"`
}

type JournalModel struct {
	DB *pgxpool.Pool
}

func (m JournalModel) AllJournals() ([]*Journal, error) {
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var journals []*Journal

	for rows.Next() {
		var journal Journal
		journal.Teacher = &User{}
		journal.Subject = &Subject{}
		err = rows.Scan(
			&journal.ID,
			&journal.Name,
			&journal.Teacher.ID,
			&journal.Teacher.Name,
			&journal.Teacher.Role,
			&journal.Subject.ID,
			&journal.Subject.Name,
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
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var journal Journal
	journal.Teacher = &User{}
	journal.Subject = &Subject{}

	err := m.DB.QueryRow(ctx, query, journalID).Scan(
		&journal.ID,
		&journal.Name,
		&journal.Teacher.ID,
		&journal.Teacher.Name,
		&journal.Teacher.Role,
		&journal.Subject.ID,
		&journal.Subject.Name,
		&journal.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
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

	err := m.DB.QueryRow(ctx, stmt, j.Name, j.Teacher.ID, j.Subject.ID, j.Archived).Scan(&j.ID)
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

	_, err := m.DB.Exec(ctx, stmt, j.Name, j.Teacher.ID, j.Subject.ID, j.Archived, j.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m JournalModel) DeleteJournal(journalID int) error {
	stmt := `DELETE FROM journals
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, journalID)
	if err != nil {
		return err
	}
	return nil
}

func (m JournalModel) GetJournalsForTeacher(teacherID int) ([]*Journal, error) {
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	WHERE teacher_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, teacherID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var journals []*Journal

	for rows.Next() {
		var journal Journal
		journal.Teacher = &User{}
		journal.Subject = &Subject{}

		err = rows.Scan(
			&journal.ID,
			&journal.Name,
			&journal.Teacher.ID,
			&journal.Teacher.Name,
			&journal.Teacher.Role,
			&journal.Subject.ID,
			&journal.Subject.Name,
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

	_, err := m.DB.Exec(ctx, stmt, userID, journalID)
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

func (m JournalModel) DeleteUserFromJournal(userID, journalID int) error {
	stmt := `DELETE FROM
	users_journals
	WHERE user_id = $1 and journal_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.Exec(ctx, stmt, userID, journalID)
	if err != nil {
		return err
	}

	affected := result.RowsAffected()

	if affected == 0 {
		return ErrUserNotInJournal
	}

	return nil
}

func (m JournalModel) GetUsersByJournalID(journalID int) ([]*User, error) {
	query := `SELECT id, name, role
	FROM users u
	INNER JOIN users_journals uj
	ON uj.user_id = u.id
	WHERE uj.journal_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID)
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
			&user.Role,
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
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	INNER JOIN users_journals uj
	ON uj.journal_id = j.id
	WHERE uj.user_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var journals []*Journal

	for rows.Next() {
		var journal Journal
		journal.Teacher = &User{}
		journal.Subject = &Subject{}

		err = rows.Scan(
			&journal.ID,
			&journal.Name,
			&journal.Teacher.ID,
			&journal.Teacher.Name,
			&journal.Teacher.Role,
			&journal.Subject.ID,
			&journal.Subject.Name,
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

func (m JournalModel) IsUserInJournal(userID, journalID int) (bool, error) {
	query := `SELECT COUNT(1) FROM users_journals
	WHERE user_id = $1 and journal_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, userID, journalID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (m JournalModel) DoesParentHaveChildInJournal(parentID, journalID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM parents_children pc
	INNER JOIN users_journals uj
	ON pc.child_id = uj.user_id
	WHERE pc.parent_id = $1
	AND uj.journal_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, parentID, journalID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}
