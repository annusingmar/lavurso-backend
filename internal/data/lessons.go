package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoSuchLesson = errors.New("no such lesson")
)

type Lesson struct {
	ID          *int       `json:"id,omitempty"`
	Journal     *Journal   `json:"journal,omitempty"`
	Subject     *Subject   `json:"subject,omitempty"`
	Description *string    `json:"description,omitempty"`
	Date        *Date      `json:"date,omitempty"`
	Course      *int       `json:"course,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Marks       []*Mark    `json:"marks,omitempty"`
}

type LessonModel struct {
	DB *pgxpool.Pool
}

// DATABASE

func (m LessonModel) InsertLesson(l *Lesson) error {
	stmt := `INSERT INTO lessons
	(journal_id, description, date, course, created_at, updated_at)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, l.Journal.ID, l.Description, l.Date.Time, l.Course, l.CreatedAt, l.UpdatedAt).Scan(&l.ID)
	if err != nil {
		return err
	}

	return nil

}

func (m LessonModel) GetLessonByID(lessonID int) (*Lesson, error) {
	query := `SELECT l.id, l.journal_id, j.name, j.archived, l.description, l.date, l.course, l.created_at, l.updated_at
	FROM lessons l
	INNER JOIN journals j
	ON j.id = l.journal_id
	WHERE l.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var lesson Lesson
	lesson.Journal = new(Journal)
	lesson.Date = new(Date)

	err := m.DB.QueryRow(ctx, query, lessonID).Scan(
		&lesson.ID,
		&lesson.Journal.ID,
		&lesson.Journal.Name,
		&lesson.Journal.Archived,
		&lesson.Description,
		&lesson.Date.Time,
		&lesson.Course,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchLesson
		default:
			return nil, err
		}
	}

	return &lesson, nil
}

func (m LessonModel) UpdateLesson(l *Lesson) error {
	stmt := `UPDATE lessons
	SET (description, date, updated_at)
	= ($1, $2, $3)
	WHERE id = $4`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, l.Description, l.Date.Time, l.UpdatedAt, l.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m LessonModel) DeleteLesson(lessonID int) error {
	stmt := `DELETE FROM lessons
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (m LessonModel) GetLessonsByJournalID(journalID int, course int) ([]*Lesson, error) {
	query := `SELECT id, journal_id, description, date, course, created_at, updated_at
	FROM lessons
	WHERE journal_id = $1 and course = $2
	ORDER BY date DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID, course)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var lessons []*Lesson

	for rows.Next() {
		var lesson Lesson
		lesson.Journal = new(Journal)
		lesson.Date = new(Date)

		err := rows.Scan(
			&lesson.ID,
			&lesson.Journal.ID,
			&lesson.Description,
			&lesson.Date.Time,
			&lesson.Course,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		lessons = append(lessons, &lesson)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (m LessonModel) GetLatestLessonsForStudent(studentID int, from, until *Date) ([]*Lesson, error) {
	sqlQuery := `SELECT l.id, s.id, s.name, l.description, l.date, l.course, l.created_at, l.updated_at
	FROM lessons l
	INNER JOIN journals j
	ON j.id = l.journal_id
	INNER JOIN users_journals uj
	ON uj.journal_id = j.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	WHERE uj.user_id = $1 AND %s
	ORDER BY date ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rows pgx.Rows
	var err error

	if until != nil {
		query := fmt.Sprintf(sqlQuery, "l.date > $2::date AND l.date <= $3::date")
		rows, err = m.DB.Query(ctx, query, studentID, from.Time, until.Time)
	} else {
		query := fmt.Sprintf(sqlQuery, "l.date > $2::date")
		rows, err = m.DB.Query(ctx, query, studentID, from.Time)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var lessons []*Lesson

	for rows.Next() {
		var lesson Lesson
		lesson.Subject = new(Subject)
		lesson.Date = new(Date)

		err := rows.Scan(
			&lesson.ID,
			&lesson.Subject.ID,
			&lesson.Subject.Name,
			&lesson.Description,
			&lesson.Date.Time,
			&lesson.Course,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		lessons = append(lessons, &lesson)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}
