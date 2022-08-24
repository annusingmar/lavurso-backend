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
	ErrNoSuchMark = errors.New("no such mark")
	ErrNoSuchType = errors.New("no such mark type")
)

type Mark struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MarkID    *int      `json:"mark_id,omitempty"`
	Lesson    *Lesson   `json:"lesson,omitempty"`
	Course    *int      `json:"course,omitempty"`
	JournalID *int      `json:"journal_id,omitempty"`
	Grade     *Grade    `json:"grade,omitempty"`
	Comment   *string   `json:"comment,omitempty"`
	Type      string    `json:"type"`
	Subject   *Subject  `json:"subject,omitempty"`
	Excuse    *Excuse   `json:"excuse,omitempty"`
	By        *User     `json:"by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		mark.Grade = new(Grade)
		mark.By = new(User)
		mark.Lesson = &Lesson{Date: new(Date)}

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
			&mark.Comment,
			&mark.Type,
			&mark.By.ID,
			&mark.By.Name,
			&mark.By.Role,
			&mark.CreatedAt,
			&mark.UpdatedAt,
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

func scanMarksWithExcuse(rows pgx.Rows) ([]*Mark, error) {
	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.Grade = new(Grade)
		mark.By = new(User)
		mark.Lesson = &Lesson{Date: new(Date)}
		mark.Excuse = &Excuse{By: new(User)}

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
			&mark.Comment,
			&mark.Type,
			&mark.By.ID,
			&mark.By.Name,
			&mark.By.Role,
			&mark.CreatedAt,
			&mark.UpdatedAt,
			&mark.Excuse.MarkID,
			&mark.Excuse.Excuse,
			&mark.Excuse.By.ID,
			&mark.Excuse.By.Name,
			&mark.Excuse.By.Role,
			&mark.Excuse.At,
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.by, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.by
	WHERE m.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var mark Mark
	mark.Grade = new(Grade)
	mark.By = new(User)
	mark.Lesson = &Lesson{Date: new(Date)}
	mark.Excuse = &Excuse{By: new(User)}

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
		&mark.Comment,
		&mark.Type,
		&mark.By.ID,
		&mark.By.Name,
		&mark.By.Role,
		&mark.CreatedAt,
		&mark.UpdatedAt,
		&mark.Excuse.MarkID,
		&mark.Excuse.Excuse,
		&mark.Excuse.By.ID,
		&mark.Excuse.By.Name,
		&mark.Excuse.By.Role,
		&mark.Excuse.At,
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

func (m MarkModel) GetMarksByStudent(userID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.by, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.by
	WHERE m.user_id = $1
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarksWithExcuse(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetLatestMarksForStudent(studentID int, from, until *Date) ([]*Mark, error) {
	sqlQuery := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, s.id, s.name, m.by, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.by, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN journals j
	ON m.journal_id = j.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	INNER JOIN users u
	ON m.by = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.by
	WHERE m.user_id = $1 AND %s
	ORDER BY m.updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rows pgx.Rows
	var err error

	if until != nil {
		query := fmt.Sprintf(sqlQuery, "m.updated_at::date > $2 AND m.updated_at::date <= $3")
		rows, err = m.DB.Query(ctx, query, studentID, from.Time, until.Time)
	} else {
		query := fmt.Sprintf(sqlQuery, "m.updated_at::date > $2")
		rows, err = m.DB.Query(ctx, query, studentID, from.Time)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.Grade = new(Grade)
		mark.By = new(User)
		mark.Lesson = &Lesson{Date: new(Date)}
		mark.Excuse = &Excuse{By: new(User)}
		mark.Subject = new(Subject)

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
			&mark.Comment,
			&mark.Type,
			&mark.Subject.ID,
			&mark.Subject.Name,
			&mark.By.ID,
			&mark.By.Name,
			&mark.By.Role,
			&mark.CreatedAt,
			&mark.UpdatedAt,
			&mark.Excuse.MarkID,
			&mark.Excuse.Excuse,
			&mark.Excuse.By.ID,
			&mark.Excuse.By.Name,
			&mark.Excuse.By.Role,
			&mark.Excuse.At,
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

func (m MarkModel) GetSubjectGradesByJournalID(journalID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
	WHERE m.journal_id = $1 and m.type = 'subject_grade'
	ORDER BY updated_at ASC`

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

func (m MarkModel) GetAllCoursesGradesByJournalID(journalID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
	WHERE m.journal_id = $1 and m.type = 'course_grade'
	ORDER BY updated_at ASC`

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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
	WHERE m.journal_id = $1 and m.course = $2 and m.type = 'course_grade'
	ORDER BY updated_at ASC`

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

func (m MarkModel) GetLessonMarksByCourseAndJournalID(journalID, course int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
	WHERE m.journal_id = $1 and m.course = $2 and m.lesson_id is not NULL
	ORDER BY updated_at ASC`

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

func (m MarkModel) GetLessonMarksForStudentByCourseAndJournalID(userID, journalID, course int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.by, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.by
	WHERE m.journal_id = $1 and m.course = $2 and m.lesson_id is not NULL and m.user_id = $3
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID, course, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	marks, err := scanMarksWithExcuse(rows)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetMarksByLessonID(lessonID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.by, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.by = u.id
	WHERE m.lesson_id = $1
	ORDER BY updated_at ASC`

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

func (m MarkModel) InsertMark(mark *Mark) error {
	stmt := `INSERT INTO marks
	(user_id, lesson_id, course, journal_id, grade_id, comment, type, by, created_at, updated_at)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, mark.UserID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.Comment, mark.Type, mark.By.ID, mark.CreatedAt, mark.UpdatedAt).Scan(&mark.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) UpdateMark(mark *Mark) error {
	stmt := `UPDATE marks
	SET (user_id, lesson_id, course, journal_id, grade_id, comment, type, by, updated_at)
	= ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	WHERE id = $10`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, mark.UserID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.Comment, mark.Type, mark.By.ID, mark.UpdatedAt, mark.ID)

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
