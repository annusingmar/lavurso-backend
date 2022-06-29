package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoSuchMark = errors.New("no such mark")
	ErrNoSuchType = errors.New("no such mark type")
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
	PreviousIDs *[]int    `json:"previous_ids,omitempty"`
	By          int       `json:"by"`
	At          time.Time `json:"at"`
}

type MarkModel struct {
	DB *sql.DB
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

func (m MarkModel) GetMarkByID(markID int) (*Mark, error) {
	query := `SELECT FROM marks
	id, user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, previous_ids, by, at
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var mark Mark

	err := m.DB.QueryRowContext(ctx, query, markID).Scan(
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
		&mark.PreviousIDs,
		&mark.By,
		&mark.At,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchMark
		default:
			return nil, err
		}
	}

	return &mark, nil
}

func (m MarkModel) InsertMark(mark *Mark) error {
	stmt := `INSERT INTO marks
	(user_id, lesson_id, course, journal_id, grade_id, subject_id, comment, type, current, previous_ids, by, at)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, mark.UserID, mark.LessonID, mark.Course, mark.JournalID, mark.GradeID,
		mark.SubjectID, mark.Comment, mark.Type, mark.Current, mark.PreviousIDs, mark.By, mark.At).Scan(&mark.ID)

	if err != nil {
		return err
	}

	return nil
}
