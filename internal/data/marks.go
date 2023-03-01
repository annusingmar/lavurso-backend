package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

var (
	ErrNoSuchMark = errors.New("no such mark")
	ErrNoSuchType = errors.New("no such mark type")
)

type Mark = model.Marks

type MarkExt struct {
	Mark
	Lesson  *Lesson    `json:"lesson,omitempty" alias:"mark_lesson"`
	Grade   *Grade     `json:"grade,omitempty"`
	Subject *Subject   `json:"subject,omitempty"`
	Excuse  *ExcuseExt `json:"excuse,omitempty"`
	Teacher *User      `json:"teacher,omitempty" alias:"teacher"`
	YearID  *int       `json:"year_id,omitempty" alias:"years.id"`
}

type MinimalMark struct {
	ID      int     `json:"id" sql:"primary_key" alias:"marks.id"`
	Type    string  `json:"type" alias:"marks.type"`
	Comment *string `json:"comment,omitempty" alias:"marks.comment"`
	Grade   *string `json:"grade,omitempty" alias:"grades.identifier"`
}

type HigherMinimalGradeMark struct {
	ID      int     `json:"id" sql:"primary_key" alias:"higher_marks.id"`
	Comment *string `json:"comment,omitempty" alias:"higher_marks.comment"`
	Grade   string  `json:"grade,omitempty" alias:"higher_marks_grade.identifier"`
}

type LessonMarks struct {
	Absent  bool `json:"absent"`
	Late    bool `json:"late"`
	NotDone bool `json:"not_done"`
}

type LessonStudent struct {
	UserExt
	Lesson LessonMarks    `json:"lesson"`
	Marks  []*MinimalMark `json:"marks,omitempty"`
}

type StudentWithLowerMarks struct {
	UserExt
	Marks      []*HigherMinimalGradeMark `json:"marks,omitempty"`
	LowerMarks []*MarkExt                `json:"lower_marks,omitempty"`
}

type MarkByLessonStudentType struct {
	LessonID  int
	StudentID int
	Type      string
}

type MarkModel struct {
	DB *sql.DB
}

const (
	MarkCommonGrade   = "grade"
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

func (m MarkModel) GetMarkAndExcuseByID(markID int) (*MarkExt, error) {
	query := postgres.SELECT(
		table.Marks.AllColumns,
		table.Excuses.AllColumns).
		FROM(table.Marks.
			LEFT_JOIN(table.Excuses, table.Excuses.MarkID.EQ(table.Marks.ID))).
		WHERE(table.Marks.ID.EQ(helpers.PostgresInt(markID)))

	var mark MarkExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &mark)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchMark
		default:
			return nil, err
		}
	}

	return &mark, nil
}

func (m MarkModel) GetMarksByStudent(userID, yearID int) ([]*MarkExt, error) {
	teacher := table.Users.AS("teacher")
	excuser := table.Users.AS("excuser")
	lesson := table.Lessons.AS("mark_lesson")

	query := postgres.SELECT(
		table.Marks.AllColumns,
		lesson.ID, lesson.Date, lesson.Description,
		table.Grades.AllColumns,
		teacher.ID, teacher.Name, teacher.Role,
		table.Excuses.AllColumns, excuser.ID, excuser.Name, excuser.Role,
	).FROM(table.Marks.
		INNER_JOIN(table.Journals, table.Journals.ID.EQ(table.Marks.JournalID)).
		LEFT_JOIN(table.Grades, table.Grades.ID.EQ(table.Marks.GradeID)).
		LEFT_JOIN(lesson, lesson.ID.EQ(table.Marks.LessonID)).
		INNER_JOIN(teacher, teacher.ID.EQ(table.Marks.TeacherID)).
		LEFT_JOIN(table.Excuses, table.Excuses.MarkID.EQ(table.Marks.ID)).
		LEFT_JOIN(excuser, excuser.ID.EQ(table.Excuses.UserID))).
		WHERE(table.Marks.UserID.EQ(helpers.PostgresInt(userID)).
			AND(table.Journals.YearID.EQ(helpers.PostgresInt(yearID)))).
		ORDER_BY(table.Marks.UpdatedAt.ASC())

	var marks []*MarkExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &marks)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetLatestMarksForStudent(studentID int, from, until *types.Date) ([]*MarkExt, error) {
	teacher := table.Users.AS("teacher")
	excuser := table.Users.AS("excuser")
	lesson := table.Lessons.AS("mark_lesson")

	where := table.Marks.UserID.EQ(helpers.PostgresInt(studentID))
	updatedAtDate := postgres.CAST(table.Marks.UpdatedAt).AS_DATE()
	if until != nil {
		where = postgres.AND(
			where,
			updatedAtDate.GT(postgres.DateT(*from.Time)),
			updatedAtDate.LT_EQ(postgres.DateT(*until.Time)),
		)
	} else {
		where = where.AND(updatedAtDate.GT(postgres.DateT(*from.Time)))
	}

	query := postgres.SELECT(
		table.Marks.AllColumns,
		lesson.ID, lesson.Date, lesson.Description,
		table.Subjects.AllColumns,
		table.Grades.AllColumns,
		teacher.ID, teacher.Name, teacher.Role,
		table.Excuses.AllColumns, excuser.ID, excuser.Name, excuser.Role,
	).FROM(table.Marks.
		INNER_JOIN(table.Journals, table.Journals.ID.EQ(table.Marks.JournalID)).
		INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID)).
		LEFT_JOIN(table.Grades, table.Grades.ID.EQ(table.Marks.GradeID)).
		LEFT_JOIN(lesson, lesson.ID.EQ(table.Marks.LessonID)).
		INNER_JOIN(teacher, teacher.ID.EQ(table.Marks.TeacherID)).
		LEFT_JOIN(table.Excuses, table.Excuses.MarkID.EQ(table.Marks.ID)).
		LEFT_JOIN(excuser, excuser.ID.EQ(table.Excuses.UserID))).
		WHERE(where).
		ORDER_BY(table.Marks.UpdatedAt.DESC())

	var marks []*MarkExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &marks)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetLessonMarksForStudentByCourseAndJournalID(userID, journalID, course int) ([]*MarkExt, error) {
	teacher := table.Users.AS("teacher")
	excuser := table.Users.AS("excuser")
	lesson := table.Lessons.AS("mark_lesson")

	query := postgres.SELECT(
		table.Marks.AllColumns,
		lesson.ID, lesson.Date, lesson.Description,
		table.Grades.AllColumns,
		teacher.ID, teacher.Name, teacher.Role,
		table.Excuses.AllColumns, excuser.ID, excuser.Name, excuser.Role,
	).FROM(table.Marks.
		LEFT_JOIN(table.Grades, table.Grades.ID.EQ(table.Marks.GradeID)).
		LEFT_JOIN(lesson, lesson.ID.EQ(table.Marks.LessonID)).
		INNER_JOIN(teacher, teacher.ID.EQ(table.Marks.TeacherID)).
		LEFT_JOIN(table.Excuses, table.Excuses.MarkID.EQ(table.Marks.ID)).
		LEFT_JOIN(excuser, excuser.ID.EQ(table.Excuses.UserID))).
		WHERE(postgres.AND(
			table.Marks.JournalID.EQ(helpers.PostgresInt(journalID)),
			table.Marks.Course.EQ(helpers.PostgresInt(course)),
			table.Marks.LessonID.IS_NOT_NULL(),
			table.Marks.UserID.EQ(helpers.PostgresInt(userID)),
		)).
		ORDER_BY(table.Marks.UpdatedAt.ASC())

	var marks []*MarkExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &marks)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) GetAllCourseSubjectGradesForStudent(studentID int) ([]*MarkExt, error) {
	query := postgres.SELECT(
		table.Marks.AllColumns,
		table.Grades.AllColumns,
		table.Subjects.AllColumns,
		table.Years.ID).
		FROM(table.Marks.
			INNER_JOIN(table.Grades, table.Grades.ID.EQ(table.Marks.GradeID)).
			INNER_JOIN(table.Journals, table.Journals.ID.EQ(table.Marks.JournalID)).
			INNER_JOIN(table.Years, table.Years.ID.EQ(table.Journals.YearID)).
			INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID))).
		WHERE(postgres.AND(
			table.Marks.Type.EQ(postgres.String(MarkCourseGrade)).OR(table.Marks.Type.EQ(postgres.String(MarkSubjectGrade))),
			table.Marks.UserID.EQ(helpers.PostgresInt(studentID)),
		)).
		ORDER_BY(
			table.Subjects.Name.ASC(),
			table.Marks.CreatedAt.ASC(),
		)

	var marks []*MarkExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &marks)
	if err != nil {
		return nil, err
	}

	return marks, nil
}

func (m MarkModel) InsertMarks(tx *sql.Tx, marks []*Mark) error {
	stmt := table.Marks.INSERT(table.Marks.MutableColumns).
		MODELS(marks).
		ON_CONFLICT(table.Marks.UserID, table.Marks.LessonID, table.Marks.Type).
		WHERE(table.Marks.Type.IN(postgres.String(MarkAbsent), postgres.String(MarkLate), postgres.String(MarkNotDone))).
		DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) UpdateMarks(tx *sql.Tx, marks []*Mark) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for _, mk := range marks {
		var ors []postgres.BoolExpression

		if mk.GradeID != nil {
			ors = append(ors, table.Marks.GradeID.IS_DISTINCT_FROM(helpers.PostgresInt(*mk.GradeID)))
		} else {
			ors = append(ors, table.Marks.GradeID.IS_NOT_NULL())
		}

		if mk.Comment != nil {
			ors = append(ors, table.Marks.Comment.IS_DISTINCT_FROM(postgres.String(*mk.Comment)))
		} else {
			ors = append(ors, table.Marks.Comment.IS_NOT_NULL())
		}

		ors = append(ors, table.Marks.Type.NOT_EQ(postgres.String(*mk.Type)))

		stmt := table.Marks.UPDATE(table.Marks.GradeID, table.Marks.Comment, table.Marks.Type, table.Marks.TeacherID, table.Marks.UpdatedAt).
			MODEL(mk).
			WHERE(table.Marks.ID.EQ(helpers.PostgresInt(mk.ID)).
				AND(postgres.OR(ors...)))

		_, err := stmt.ExecContext(ctx, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m MarkModel) DeleteMarks(tx *sql.Tx, markIDs []int) error {
	var mids []postgres.Expression
	for _, mid := range markIDs {
		mids = append(mids, helpers.PostgresInt(mid))
	}

	stmt := table.Marks.DELETE().
		WHERE(table.Marks.ID.IN(mids...))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) DeleteMarksByStudentIDType(tx *sql.Tx, l []MarkByLessonStudentType) error {
	var or []postgres.BoolExpression
	for _, m := range l {
		or = append(or, postgres.AND(
			table.Marks.LessonID.EQ(helpers.PostgresInt(m.LessonID)),
			table.Marks.UserID.EQ(helpers.PostgresInt(m.StudentID)),
			table.Marks.Type.EQ(postgres.String(m.Type)),
		))
	}

	stmt := table.Marks.DELETE().
		WHERE(postgres.OR(or...))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

func (m MarkModel) GetStudentsMarksForLesson(lessonID int) ([]*LessonStudent, error) {
	query := postgres.SELECT(
		table.Users.ID, table.Users.Name, table.Marks.ID,
		postgres.CASE().
			WHEN(table.Marks.Type.EQ(postgres.String(MarkLessonGrade))).
			THEN(postgres.String("grade")).
			ELSE(table.Marks.Type).AS("marks.type"),
		table.Marks.Comment, table.Grades.Identifier,
	).
		FROM(table.Lessons.
			INNER_JOIN(table.StudentsJournals, table.StudentsJournals.JournalID.EQ(table.Lessons.JournalID)).
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.StudentsJournals.StudentID)).
			LEFT_JOIN(table.Marks, table.Marks.UserID.EQ(table.Users.ID).AND(table.Marks.LessonID.EQ(table.Lessons.ID))).
			LEFT_JOIN(table.Grades, table.Grades.ID.EQ(table.Marks.GradeID))).
		WHERE(table.Lessons.ID.EQ(helpers.PostgresInt(lessonID))).
		ORDER_BY(table.Users.Name.ASC(), table.Marks.CreatedAt.ASC())

	var students []*LessonStudent

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &students)
	if err != nil {
		return nil, err
	}

	// todo: can be improved?
	for _, s := range students {
		marks := make([]*MinimalMark, 0, len(s.Marks))
		for _, m := range s.Marks {
			switch m.Type {
			case MarkAbsent:
				s.Lesson.Absent = true
			case MarkLate:
				s.Lesson.Late = true
			case MarkNotDone:
				s.Lesson.NotDone = true
			default:
				marks = append(marks, m)
			}
		}
		s.Marks = marks
	}

	return students, nil
}

func (m MarkModel) GetStudentsMarksForCourse(journalID, course int) ([]*StudentWithLowerMarks, error) {
	courseMarks := table.Marks.AS("higher_marks")
	courseMarksGrade := table.Grades.AS("higher_marks_grade")
	lessonMarks := table.Marks.AS("marks")
	lessonMarksGrade := table.Grades.AS("grades")
	lessonMarksTeacher := table.Users.AS("teacher")
	lesson := table.Lessons.AS("mark_lesson")

	query := postgres.SELECT(
		table.Users.ID, table.Users.Name,
		courseMarks.ID, courseMarks.Comment, courseMarksGrade.Identifier,
		lessonMarks.AllColumns,
		lessonMarksGrade.Identifier, lessonMarksGrade.Value,
		lessonMarksTeacher.ID, lessonMarksTeacher.Name, lessonMarksTeacher.Role,
		lesson.ID, lesson.Date, lesson.Description,
		table.Excuses.MarkID,
	).
		FROM(table.Journals.
			INNER_JOIN(table.StudentsJournals, table.StudentsJournals.JournalID.EQ(helpers.PostgresInt(journalID))).
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.StudentsJournals.StudentID)).
			LEFT_JOIN(courseMarks, postgres.AND(
				courseMarks.UserID.EQ(table.Users.ID),
				courseMarks.JournalID.EQ(helpers.PostgresInt(journalID)),
				courseMarks.Course.EQ(helpers.PostgresInt(course)),
				courseMarks.Type.EQ(postgres.String(MarkCourseGrade)),
			)).
			LEFT_JOIN(courseMarksGrade, courseMarksGrade.ID.EQ(courseMarks.GradeID)).
			LEFT_JOIN(lessonMarks, postgres.AND(
				lessonMarks.UserID.EQ(table.Users.ID),
				lessonMarks.JournalID.EQ(helpers.PostgresInt(journalID)),
				lessonMarks.Course.EQ(helpers.PostgresInt(course)),
				lessonMarks.LessonID.IS_NOT_NULL(),
			)).
			LEFT_JOIN(lessonMarksGrade, lessonMarksGrade.ID.EQ(lessonMarks.GradeID)).
			LEFT_JOIN(lessonMarksTeacher, lessonMarksTeacher.ID.EQ(lessonMarks.TeacherID)).
			LEFT_JOIN(lesson, lesson.ID.EQ(lessonMarks.LessonID)).
			LEFT_JOIN(table.Excuses, table.Excuses.MarkID.EQ(lessonMarks.ID))).
		ORDER_BY(table.Users.Name.ASC(), courseMarks.CreatedAt.ASC(), lesson.Date.DESC())

	var students []*StudentWithLowerMarks

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &students)
	if err != nil {
		return nil, err
	}

	return students, nil

}

func (m MarkModel) GetStudentsMarksForJournalSubject(journalID int) ([]*StudentWithLowerMarks, error) {
	subjectMarks := table.Marks.AS("higher_marks")
	subjectMarksGrade := table.Grades.AS("higher_marks_grade")
	courseMarks := table.Marks.AS("marks")
	courseMarksGrade := table.Grades.AS("grades")
	courseMarksTeacher := table.Users.AS("teacher")

	query := postgres.SELECT(
		table.Users.ID, table.Users.Name,
		subjectMarks.ID, subjectMarks.Comment, subjectMarksGrade.Identifier,
		courseMarks.AllColumns,
		courseMarksGrade.Identifier, courseMarksGrade.Value,
		courseMarksTeacher.ID, courseMarksTeacher.Name, courseMarksTeacher.Role,
	).
		FROM(table.Journals.
			INNER_JOIN(table.StudentsJournals, table.StudentsJournals.JournalID.EQ(helpers.PostgresInt(journalID))).
			INNER_JOIN(table.Users, table.Users.ID.EQ(table.StudentsJournals.StudentID)).
			LEFT_JOIN(subjectMarks, postgres.AND(
				subjectMarks.UserID.EQ(table.Users.ID),
				subjectMarks.JournalID.EQ(helpers.PostgresInt(journalID)),
				subjectMarks.Type.EQ(postgres.String(MarkSubjectGrade)),
			)).
			LEFT_JOIN(subjectMarksGrade, subjectMarksGrade.ID.EQ(subjectMarks.GradeID)).
			LEFT_JOIN(courseMarks, postgres.AND(
				courseMarks.UserID.EQ(table.Users.ID),
				courseMarks.JournalID.EQ(helpers.PostgresInt(journalID)),
				courseMarks.Type.EQ(postgres.String(MarkCourseGrade)),
			)).
			LEFT_JOIN(courseMarksGrade, courseMarksGrade.ID.EQ(courseMarks.GradeID)).
			LEFT_JOIN(courseMarksTeacher, courseMarksTeacher.ID.EQ(courseMarks.TeacherID))).
		ORDER_BY(table.Users.Name.ASC(), subjectMarks.CreatedAt.ASC(), courseMarks.UpdatedAt.DESC())

	var students []*StudentWithLowerMarks

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &students)
	if err != nil {
		return nil, err
	}

	return students, nil
}

func (m MarkModel) GetMarkIDsForLesson(lessonID int) ([]int, error) {
	query := postgres.SELECT(table.Marks.ID).
		FROM(table.Marks).
		WHERE(table.Marks.LessonID.EQ(helpers.PostgresInt(lessonID)).
			AND(table.Marks.Type.NOT_IN(postgres.String(MarkAbsent), postgres.String(MarkLate), postgres.String(MarkNotDone))))

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m MarkModel) GetMarkIDsForCourse(journalID, course int) ([]int, error) {
	query := postgres.SELECT(table.Marks.ID).
		FROM(table.Marks).
		WHERE(postgres.AND(
			table.Marks.JournalID.EQ(helpers.PostgresInt(journalID)),
			table.Marks.Course.EQ(helpers.PostgresInt(course)),
			table.Marks.Type.EQ(postgres.String(MarkCourseGrade)),
		))

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m MarkModel) GetMarkIDsForJournalSubject(journalID int) ([]int, error) {
	query := postgres.SELECT(table.Marks.ID).
		FROM(table.Marks).
		WHERE(postgres.AND(
			table.Marks.JournalID.EQ(helpers.PostgresInt(journalID)),
			table.Marks.Type.EQ(postgres.String(MarkSubjectGrade)),
		))

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
