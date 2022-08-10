package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNotValidAbsence = errors.New("not valid absence for user")
	ErrNoSuchAbsence   = errors.New("no such absence")
	ErrAbsenceExcused  = errors.New("absence already excused")
	ErrNoSuchExcuse    = errors.New("no such excuse")
)

type AbsenceExcuse struct {
	ID            *int       `json:"id"`
	AbsenceMarkID *int       `json:"absence_id"`
	Excuse        *string    `json:"excuse"`
	By            *int       `json:"by"`
	At            *time.Time `json:"at"`
}

type AbsenceModel struct {
	DB *pgxpool.Pool
}

func (m AbsenceModel) GetAbsenceMarksByUserID(userID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, l.date, l.description, m.course, m.journal_id, m.grade_id, m.subject_id, m.comment, m.type, m.by, m.at, exc.id, exc.absence_mark_id, exc.excuse, exc.by, exc.at
	FROM marks m
	LEFT JOIN absences_excuses exc
	ON m.id = exc.absence_mark_id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.user_id = $1 and m.type = 'absent'`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.AbsenceExcuses = new(AbsenceExcuse)
		mark.Lesson = &Lesson{Date: &Date{}}

		err := rows.Scan(
			&mark.ID,
			&mark.UserID,
			&mark.Lesson.ID,
			&mark.Lesson.Date.Time,
			&mark.Lesson.Description,
			&mark.Course,
			&mark.JournalID,
			&mark.Grade.ID,
			&mark.SubjectID,
			&mark.Comment,
			&mark.Type,
			&mark.By,
			&mark.At,
			&mark.AbsenceExcuses.ID,
			&mark.AbsenceExcuses.AbsenceMarkID,
			&mark.AbsenceExcuses.Excuse,
			&mark.AbsenceExcuses.By,
			&mark.AbsenceExcuses.At,
		)
		if err != nil {
			return nil, err
		}

		if mark.AbsenceExcuses.AbsenceMarkID == nil {
			mark.AbsenceExcuses = nil
		}

		marks = append(marks, &mark)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return marks, nil
}

func (m AbsenceModel) GetAbsenceByMarkID(markID int) (*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, l.date, l.description, m.course, m.journal_id, m.grade_id, m.subject_id, m.comment, m.type, m.by, m.at, exc.id, exc.absence_mark_id, exc.excuse, exc.by, exc.at
	FROM marks m
	LEFT JOIN absences_excuses exc
	ON m.id = exc.absence_mark_id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.id = $1 and m.type = 'absent'`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var mark Mark
	mark.AbsenceExcuses = new(AbsenceExcuse)

	err := m.DB.QueryRow(ctx, query, markID).Scan(
		&mark.ID,
		&mark.UserID,
		&mark.Lesson.ID,
		&mark.Lesson.Date.Time,
		&mark.Lesson.Description,
		&mark.Course,
		&mark.JournalID,
		&mark.Grade.ID,
		&mark.SubjectID,
		&mark.Comment,
		&mark.Type,
		&mark.By,
		&mark.At,
		&mark.AbsenceExcuses.ID,
		&mark.AbsenceExcuses.AbsenceMarkID,
		&mark.AbsenceExcuses.Excuse,
		&mark.AbsenceExcuses.By,
		&mark.AbsenceExcuses.At,
	)

	if mark.AbsenceExcuses.AbsenceMarkID == nil {
		mark.AbsenceExcuses = nil
	}

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchAbsence
		default:
			return nil, err
		}
	}

	return &mark, nil
}

func (m AbsenceModel) InsertExcuse(excuse *AbsenceExcuse) error {
	stmt := `INSERT INTO absences_excuses
	(absence_mark_id, excuse, by, at)
	VALUES
	($1, $2, $3, $4)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, excuse.AbsenceMarkID, excuse.Excuse, excuse.By, excuse.At).Scan(&excuse.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m AbsenceModel) DeleteExcuse(excuseID int) error {
	stmt := `DELETE FROM absences_excuses
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, excuseID)
	if err != nil {
		return err
	}

	return nil
}

func (m AbsenceModel) GetExcuseByID(excuseID int) (*AbsenceExcuse, error) {
	query := `SELECT id, absence_mark_id, excuse, by, at
	FROM absences_excuses
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var excuse AbsenceExcuse

	err := m.DB.QueryRow(ctx, query, excuseID).Scan(
		&excuse.ID,
		&excuse.AbsenceMarkID,
		&excuse.Excuse,
		&excuse.By,
		&excuse.At,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchExcuse
		default:
			return nil, err
		}
	}

	return &excuse, nil
}
