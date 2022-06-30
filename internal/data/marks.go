package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
)

var (
	ErrNoSuchMark     = errors.New("no such mark")
	ErrNoSuchType     = errors.New("no such mark type")
	ErrMarkNotCurrent = errors.New("mark is not current")
	ErrMarkDeleted    = errors.New("mark is deleted")
)

type Mark struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	LessonID    *int      `json:"lesson_id,omitempty"`
	Course      *int      `json:"course,omitempty"`
	JournalID   *int      `json:"journal_id,omitempty"`
	GradeID     *int      `json:"grade_id,omitempty"`
	SubjectID   *int      `json:"subject_id,omitempty"`
	Comment     *string   `json:"comment,omitempty"`
	Type        string    `json:"type"`
	Current     bool      `json:"current"`
	Deleted     bool      `json:"deleted"`
	PreviousIDs []int     `json:"previous_ids,omitempty"`
	By          int       `json:"by"`
	At          time.Time `json:"at"`
}

type MarkModel struct {
	DB *pgx.Conn
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

		err := rows.Scan(
			&mark.ID,
			&mark.UserID,
			&mark.LessonID,
			&mark.Course,
			&mark.JournalID,
			&mark.GradeID,
			&mark.SubjectID,
			&mark.Comment,
			&mark.Type,
			&mark.Current,
			&mark.Deleted,
			&mark.PreviousIDs,
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
	id, user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, deleted, previous_ids, by, at
	FROM marks
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var mark Mark

	err := m.DB.QueryRow(ctx, query, markID).Scan(
		&mark.ID,
		&mark.UserID,
		&mark.LessonID,
		&mark.Course,
		&mark.JournalID,
		&mark.GradeID,
		&mark.SubjectID,
		&mark.Comment,
		&mark.Type,
		&mark.Current,
		&mark.Deleted,
		&mark.PreviousIDs,
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
	id, user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, deleted, previous_ids, by, at
	FROM marks
	WHERE user_id = $1`

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
	id, user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, deleted, previous_ids, by, at
	FROM marks
	WHERE journal_id = $1`

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

func (m MarkModel) GetMarksByUserIDAndJournalID(userID, journalID int) ([]*Mark, error) {
	query := `SELECT
	id, user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, deleted, previous_ids, by, at
	FROM marks
	WHERE user_id = $1 and journal_id = $2`

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
	(user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, previous_ids, by, at)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, mark.UserID, mark.LessonID, mark.Course, mark.JournalID, mark.GradeID,
		mark.SubjectID, mark.Comment, mark.Type, mark.Current, mark.PreviousIDs, mark.By, mark.At).Scan(&mark.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) SetMarkNotCurrent(markID, by int) error {
	stmt := `UPDATE marks
	SET (current, by, at)
	= ($1, $2, $3)
	WHERE id = $4`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, false, by, time.Now().UTC(), markID)
	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) UpdateMark(mark *Mark) error {
	stmt := `UPDATE marks
	SET (user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, deleted, previous_ids, by, at)
	= ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	WHERE id = $14`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, mark.UserID, mark.LessonID, mark.Course, mark.JournalID, mark.GradeID,
		mark.SubjectID, mark.Comment, mark.Type, mark.Current, mark.Deleted, mark.PreviousIDs, mark.By, mark.At, mark.ID)

	if err != nil {
		return err
	}

	return nil
}
