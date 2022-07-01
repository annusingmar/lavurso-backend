package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNoSuchLesson = errors.New("no such lesson")
)

type Lesson struct {
	ID          int       `json:"id"`
	JournalID   int       `json:"journal_id"`
	Description string    `json:"description"`
	Date        Date      `json:"date"`
	Course      int       `json:"course"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int       `json:"version"`
}

type LessonModel struct {
	DB *pgxpool.Pool
}

// DATABASE

func (m LessonModel) InsertLesson(l *Lesson) error {
	stmt := `INSERT INTO lessons
	(journal_id, description, date, course, created_at, updated_at, version)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, l.JournalID, l.Description, l.Date.Time, l.Course, l.CreatedAt, l.UpdatedAt, l.Version).Scan(&l.ID)
	if err != nil {
		return err
	}

	return nil

}

func (m LessonModel) GetLessonByID(lessonID int) (*Lesson, error) {
	query := `SELECT id, journal_id, description, date, course, created_at, updated_at, version
	FROM lessons
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var lesson Lesson

	err := m.DB.QueryRow(ctx, query, lessonID).Scan(
		&lesson.ID,
		&lesson.JournalID,
		&lesson.Description,
		&lesson.Date.Time,
		&lesson.Course,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
		&lesson.Version,
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
	SET (description, date, updated_at, version)
	= ($1, $2, $3, version+1)
	WHERE id = $4 and version = $5
	RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, l.Description, l.Date.Time, l.UpdatedAt, l.ID, l.Version).Scan(&l.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m LessonModel) GetLessonsByJournalID(journalID int) ([]*Lesson, error) {
	query := `SELECT id, journal_id, description, date, course, created_at, updated_at, version
	FROM lessons
	WHERE journal_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var lessons []*Lesson

	for rows.Next() {
		var lesson Lesson
		err := rows.Scan(
			&lesson.ID,
			&lesson.JournalID,
			&lesson.Description,
			&lesson.Date.Time,
			&lesson.Course,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
			&lesson.Version,
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
