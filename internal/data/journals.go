package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoSuchJournal        = errors.New("no such journal")
	ErrUserAlreadyInJournal = errors.New("user is already part of journal")
	ErrJournalArchived      = errors.New("journal is archived")
	ErrJournalNotArchived   = errors.New("journal is not archived")
	ErrUserNotInJournal     = errors.New("user not in journal")
)

type Journal struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Teacher       *User      `json:"teacher,omitempty"`
	Subject       *Subject   `json:"subject,omitempty"`
	LastUpdated   *time.Time `json:"last_updated,omitempty"`
	CurrentCourse *int       `json:"current_course,omitempty"`
	Archived      *bool      `json:"archived,omitempty"`
}

type JournalModel struct {
	DB *pgxpool.Pool
}

func (m JournalModel) AllJournals(archived bool) ([]*Journal, error) {
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.last_updated, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	WHERE j.archived = $1
	ORDER BY j.last_updated DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, archived)
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
			&journal.LastUpdated,
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
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.last_updated, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	WHERE j.id = $1`

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
		&journal.LastUpdated,
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
	SET (name, teacher_id, subject_id, last_updated, archived)
	= ($1, $2, $3, $4, $5)
	WHERE id = $6`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, j.Name, j.Teacher.ID, j.Subject.ID, time.Now().UTC(), j.Archived, j.ID)
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

func (m JournalModel) GetJournalsForTeacher(teacherID int, archived bool) ([]*Journal, error) {
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.last_updated, j.archived
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	WHERE j.teacher_id = $1 and j.archived = $2
	ORDER BY j.last_updated DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, teacherID, archived)
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
			&journal.LastUpdated,
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
	ORDER BY name ASC`

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
	query := `SELECT j.id, j.name, j.teacher_id, u.name, u.role, j.subject_id, s.name, j.last_updated, j.archived
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
			&journal.LastUpdated,
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

func (m JournalModel) SetJournalLastUpdated(journalID int) error {
	stmt := `UPDATE journals
	SET last_updated = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, time.Now().UTC(), journalID)
	if err != nil {
		return err
	}

	return nil
}

func (m JournalModel) GetCurrentCourseForJournal(journalID int) (*int, error) {
	query := `SELECT coalesce(max(l.course),1) AS course
	FROM lessons l
	INNER JOIN journals j
	ON l.journal_id = j.id
	WHERE l.journal_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var course *int

	err := m.DB.QueryRow(ctx, query, journalID).Scan(&course)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (m JournalModel) SetJournalArchived(journalID int, archived bool) error {
	stmt := `UPDATE journals
	SET archived = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, archived, journalID)
	if err != nil {
		return err
	}

	return nil
}
