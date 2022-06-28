package data

import (
	"context"
	"database/sql"
	"time"
)

type Lesson struct {
	ID          int       `json:"id"`
	JournalID   int       `json:"journal_id"`
	Description string    `json:"description"`
	Date        Date      `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int       `json:"version"`
}

type LessonModel struct {
	DB *sql.DB
}

// DATABASE

func (m LessonModel) InsertLesson(l *Lesson) error {
	stmt := `INSERT INTO lessons
	(journal_id, description, date, created_at, updated_at, version)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, l.JournalID, l.Description, l.Date.Time, l.CreatedAt, l.UpdatedAt, l.Version).Scan(&l.ID)
	if err != nil {
		return err
	}

	return nil

}
