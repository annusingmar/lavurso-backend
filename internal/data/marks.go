package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNoSuchMark = errors.New("no such mark")
	ErrNoSuchType = errors.New("no such mark type")
)

type Mark struct {
	ID             int            `json:"id"`
	UserID         int            `json:"user_id"`
	MarkID         *int           `json:"mark_id,omitempty"`
	Lesson         *Lesson        `json:"lesson,omitempty"`
	Course         *int           `json:"course,omitempty"`
	JournalID      *int           `json:"journal_id,omitempty"`
	Grade          *Grade         `json:"grade,omitempty"`
	SubjectID      *int           `json:"subject_id,omitempty"`
	Comment        *string        `json:"comment,omitempty"`
	Type           string         `json:"type"`
	AbsenceExcuses *AbsenceExcuse `json:"absence_excuses,omitempty"`
	By             int            `json:"by"`
	At             time.Time      `json:"at"`
}

type MarkModel struct {
	DB *pgxpool.Pool
}

const (
	MarkLessonGrade   = "lesson_grade"
	MarkCourseGrade   = "course_grade"
	MarkSubjectGrade  = "subject_grade"
	MarkNotDone       = "not_done"
	MarkNoticeGood    = "notice_good"
	MarkNoticeNeutral = "notice_neutral"
	MarkNoticeBad     = "notice_bad"
	MarkAbsent        = "absent"
	MarkLate          = "late"
)

func scanMarks(rows pgx.Rows) ([]*Mark, error) {
	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.Grade = &Grade{}
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
			&mark.Grade.Identifier,
			&mark.Grade.Value,
			&mark.SubjectID,
			&mark.Comment,
			&mark.Type,
			&mark.By,
			&mark.At,
		)
		if err != nil {
			return nil, err
		}

		marks = append(marks, &mark)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetMarkByID(markID int) (*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var mark Mark
	mark.Grade = &Grade{}
	mark.Lesson = &Lesson{Date: &Date{}}

	err := m.DB.QueryRow(ctx, query, markID).Scan(
		&mark.ID,
		&mark.UserID,
		&mark.Lesson.ID,
		&mark.Lesson.Date.Time,
		&mark.Lesson.Description,
		&mark.Course,
		&mark.JournalID,
		&mark.Grade.ID,
		&mark.Grade.Identifier,
		&mark.Grade.Value,
		&mark.SubjectID,
		&mark.Comment,
		&mark.Type,
		&mark.By,
		&mark.At,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchMark
		default:
			return nil, err
		}
	}

	return &mark, nil
}

func (m MarkModel) GetMarksByUserID(userID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE user_id = $1
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetMarksByJournalID(journalID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE journal_id = $1
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetSubjectGradesByJournalID(journalID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.journal_id = $1 and m.type = 'subject_grade'
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetCourseGradesByJournalID(journalID, course int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.journal_id = $1 and m.course = $2 and m.type = 'course_grade'
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID, course)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetLessonGradesByJournalID(journalID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.journal_id = $1 and m.type = 'lesson_grade'
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetMarksByLessonID(lessonID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE m.lesson_id = $1
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, lessonID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetMarksByUserIDAndJournalID(userID, journalID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE user_id = $1 and journal_id = $2
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) InsertMark(mark *Mark) error {
	stmt := `INSERT INTO marks
	(user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, by, at)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, mark.UserID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.SubjectID, mark.Comment, mark.Type, mark.By, mark.At).Scan(&mark.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) UpdateMark(mark *Mark) error {
	stmt := `UPDATE marks
	SET (user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, by, at)
	= ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	WHERE id = $11`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, mark.UserID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.SubjectID, mark.Comment, mark.Type, mark.By, mark.At, mark.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) DeleteMark(markID int) error {
	stmt := `DELETE FROM marks
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, markID)
	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) InsertOldMark(mark *Mark) error {
	stmt := `INSERT INTO mark_history
	(user_id, mark_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, by, at)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, mark.UserID, mark.MarkID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.SubjectID, mark.Comment, mark.Type, mark.By, mark.At).Scan(&mark.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) GetOldMarksByMarkID(markID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.subject_id, m.comment, m.type, m.by, m.at
	FROM mark_history m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	WHERE mark_id = $1
	ORDER BY at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, markID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarks(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}
