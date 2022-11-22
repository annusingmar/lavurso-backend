package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

var (
	ErrNoSuchLesson = errors.New("no such lesson")
)

type Lesson struct {
	ID          *int        `json:"id,omitempty"`
	Journal     *Journal    `json:"journal,omitempty"`
	Subject     *Subject    `json:"subject,omitempty"`
	Description *string     `json:"description,omitempty"`
	Date        *types.Date `json:"date,omitempty"`
	Course      *int        `json:"course,omitempty"`
	CreatedAt   *time.Time  `json:"created_at,omitempty"`
	UpdatedAt   *time.Time  `json:"updated_at,omitempty"`
	Marks       []*Mark     `json:"marks,omitempty"`
}

type NLesson struct {
	model.Lessons
	Journal *model.Journals `json:"journal,omitempty"`
	Subject *model.Subjects `json:"subject,omitempty"`
	Marks   []*NMark        `json:"marks,omitempty"`
}

type LessonModel struct {
	DB *sql.DB
}

// DATABASE

func (m LessonModel) InsertLesson(l *NLesson) error {
	stmt := table.Lessons.INSERT(table.Lessons.MutableColumns).
		MODEL(l).
		RETURNING(table.Lessons.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, &l)
	if err != nil {
		return err
	}

	return nil

}

func (m LessonModel) GetLessonByID(lessonID int) (*NLesson, error) {
	query := postgres.SELECT(table.Lessons.AllColumns, table.Journals.AllColumns).
		FROM(table.Lessons.
			INNER_JOIN(table.Journals, table.Journals.ID.EQ(table.Lessons.JournalID))).
		WHERE(table.Lessons.ID.EQ(postgres.Int32(int32(lessonID))))

	var lesson NLesson

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &lesson)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchLesson
		default:
			return nil, err
		}
	}

	return &lesson, nil
}

func (m LessonModel) UpdateLesson(l *NLesson) error {
	stmt := table.Lessons.UPDATE(table.Lessons.Description, table.Lessons.Date, table.Lessons.UpdatedAt).
		MODEL(l).
		WHERE(table.Lessons.ID.EQ(postgres.Int32(int32(l.ID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m LessonModel) DeleteLesson(lessonID int) error {
	stmt := table.Lessons.DELETE().WHERE(table.Lessons.ID.EQ(postgres.Int32(int32(lessonID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m LessonModel) GetLessonsByJournalID(journalID int, course int) ([]*NLesson, error) {
	query := postgres.SELECT(table.Lessons.AllColumns).
		FROM(table.Lessons).
		WHERE(table.Lessons.JournalID.EQ(postgres.Int32(int32(journalID))).
			AND(table.Lessons.Course.EQ(postgres.Int32(int32(course))))).
		ORDER_BY(table.Lessons.Date.DESC())

	var lessons []*NLesson

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &lessons)
	if err != nil {
		return nil, err
	}

	return lessons, nil
}

func (m LessonModel) GetLessonsAndStudentMarksByJournalID(studentID, journalID, course int) ([]*NLesson, error) {
	teacher := table.Users.AS("teacher")
	excuser := table.Users.AS("excuser")

	query := postgres.SELECT(
		table.Lessons.AllColumns,
		table.Marks.AllColumns,
		table.Grades.AllColumns,
		table.Excuses.AllColumns,
		teacher.ID, teacher.Name, teacher.Role,
		excuser.ID, excuser.Name, excuser.Role).
		FROM(table.Lessons.
			LEFT_JOIN(table.Marks, table.Marks.LessonID.EQ(table.Lessons.ID).
				AND(table.Marks.UserID.EQ(postgres.Int32(int32(studentID))))).
			LEFT_JOIN(table.Grades, table.Grades.ID.EQ(table.Marks.GradeID)).
			LEFT_JOIN(table.Excuses, table.Excuses.MarkID.EQ(table.Marks.ID)).
			LEFT_JOIN(teacher, teacher.ID.EQ(table.Marks.TeacherID)).
			LEFT_JOIN(excuser, excuser.ID.EQ(table.Excuses.UserID))).
		WHERE(postgres.AND(
			table.Lessons.JournalID.EQ(postgres.Int32(int32(journalID))),
			table.Lessons.Course.EQ(postgres.Int32(int32(course))),
		)).
		ORDER_BY(table.Lessons.Date.DESC(), table.Marks.UpdatedAt.ASC())

	var lessons []*NLesson

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &lessons)
	if err != nil {
		return nil, err
	}

	return lessons, nil
}

func (m LessonModel) GetLatestLessonsForStudent(studentID int, from, until *types.Date) ([]*NLesson, error) {
	query := postgres.SELECT(table.Lessons.AllColumns, table.Subjects.AllColumns).
		FROM(table.Lessons.
			INNER_JOIN(table.Journals, table.Journals.ID.EQ(table.Lessons.JournalID)).
			INNER_JOIN(table.UsersJournals, table.UsersJournals.JournalID.EQ(table.Journals.ID)).
			INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID)))

	if until != nil {
		query = query.WHERE(postgres.AND(
			table.UsersJournals.UserID.EQ(postgres.Int32(int32(studentID))),
			table.Lessons.Date.GT(postgres.DateT(*from.Time)),
			table.Lessons.Date.LT_EQ(postgres.DateT(*until.Time)),
		)).ORDER_BY(table.Lessons.Date.DESC())
	} else {
		query = query.WHERE(postgres.AND(
			table.UsersJournals.UserID.EQ(postgres.Int32(int32(studentID))),
			table.Lessons.Date.GT(postgres.DateT(*from.Time)),
		)).ORDER_BY(table.Lessons.Date.DESC())
	}

	var lessons []*NLesson

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &lessons)
	if err != nil {
		return nil, err
	}

	return lessons, nil
}
