package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-jet/jet/v2/postgres"
)

var (
	ErrNoSuchMark = errors.New("no such mark")
	ErrNoSuchType = errors.New("no such mark type")
)

type Mark struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Lesson    *Lesson   `json:"lesson,omitempty"`
	Course    *int      `json:"course,omitempty"`
	JournalID *int      `json:"journal_id,omitempty"`
	Grade     *Grade    `json:"grade,omitempty"`
	Comment   *string   `json:"comment,omitempty"`
	Type      string    `json:"type"`
	Subject   *Subject  `json:"subject,omitempty"`
	Excuse    *Excuse   `json:"excuse,omitempty"`
	Teacher   *User     `json:"teacher"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NMark struct {
	model.Marks
	Lesson  *model.Lessons  `json:"lesson,omitempty" alias:"mark_lesson"`
	Grade   *model.Grades   `json:"grade,omitempty"`
	Subject *model.Subjects `json:"subject,omitempty"`
	Excuse  *NExcuse        `json:"excuse,omitempty"`
	Teacher *model.Users    `json:"teacher,omitempty" alias:"teacher"`
}

type MinimalMark struct {
	ID      int     `json:"id" sql:"primary_key" alias:"marks.id"`
	Type    string  `json:"type" alias:"marks.type"`
	Comment *string `json:"comment,omitempty" alias:"marks.comment"`
	Grade   *string `json:"grade,omitempty" alias:"grades.identifier"`
}

type LessonMarks struct {
	Absent  bool           `json:"absent" alias:"absent"`
	Late    bool           `json:"late" alias:"late"`
	NotDone bool           `json:"not_done" alias:"not_done"`
	Marks   []*MinimalMark `json:"marks,omitempty" alias:"marks"`
}

type LessonStudent struct {
	NUser
	Lesson LessonMarks `json:"lesson" alias:"lesson"`
}

type MarkByStudentIDType struct {
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

func scanMarks(rows *sql.Rows) ([]*Mark, error) {
	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.Grade = new(Grade)
		mark.Teacher = new(User)
		mark.Lesson = &Lesson{Date: new(types.Date)}

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
			&mark.Teacher.ID,
			&mark.Teacher.Name,
			&mark.Teacher.Role,
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

func scanMarksWithExcuse(rows *sql.Rows) ([]*Mark, error) {
	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.Grade = new(Grade)
		mark.Teacher = new(User)
		mark.Lesson = &Lesson{Date: new(types.Date)}
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
			&mark.Teacher.ID,
			&mark.Teacher.Name,
			&mark.Teacher.Role,
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.user_id, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.teacher_id = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.user_id
	WHERE m.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var mark Mark
	mark.Grade = new(Grade)
	mark.Teacher = new(User)
	mark.Lesson = &Lesson{Date: new(types.Date)}
	mark.Excuse = &Excuse{By: new(User)}

	err := m.DB.QueryRowContext(ctx, query, markID).Scan(
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
		&mark.Teacher.ID,
		&mark.Teacher.Name,
		&mark.Teacher.Role,
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
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchMark
		default:
			return nil, err
		}
	}

	return &mark, nil
}

func (m MarkModel) GetMarksByStudent(userID, yearID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.user_id, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	LEFT JOIN journals j
	ON m.journal_id = j.id
	INNER JOIN users u
	ON m.teacher_id = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.user_id
	WHERE m.user_id = $1 and j.year_id = $2
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID, yearID)
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

func (m MarkModel) GetLatestMarksForStudent(studentID int, from, until *types.Date) ([]*Mark, error) {
	sqlQuery := `SELECT
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, s.id, s.name, m.teacher_id, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.user_id, u2.name, u2.role, ex.at
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
	ON m.teacher_id = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.user_id
	WHERE m.user_id = $1 AND %s
	ORDER BY m.updated_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rows *sql.Rows
	var err error

	if until != nil {
		query := fmt.Sprintf(sqlQuery, "m.updated_at::date > $2::date AND m.updated_at::date <= $3::date")
		rows, err = m.DB.QueryContext(ctx, query, studentID, from.Time, until.Time)
	} else {
		query := fmt.Sprintf(sqlQuery, "m.updated_at::date > $2::date")
		rows, err = m.DB.QueryContext(ctx, query, studentID, from.Time)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.Grade = new(Grade)
		mark.Teacher = new(User)
		mark.Lesson = &Lesson{Date: new(types.Date)}
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
			&mark.Teacher.ID,
			&mark.Teacher.Name,
			&mark.Teacher.Role,
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.teacher_id = u.id
	WHERE m.journal_id = $1 and m.type = 'subject_grade'
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID)
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.teacher_id = u.id
	WHERE m.journal_id = $1 and m.type = 'course_grade'
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID)
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.teacher_id = u.id
	WHERE m.journal_id = $1 and m.course = $2 and m.type = 'course_grade'
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID, course)
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.teacher_id = u.id
	WHERE m.journal_id = $1 and m.course = $2 and m.lesson_id is not NULL
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID, course)
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
	m.id, m.user_id, m.lesson_id, l.date, l.description, m.course, m.journal_id, m.grade_id, g.identifier, g.value, m.comment, m.type, m.teacher_id, u.name, u.role, m.created_at, m.updated_at, ex.mark_id, ex.excuse, ex.user_id, u2.name, u2.role, ex.at
	FROM marks m
	LEFT JOIN grades g
	ON m.grade_id = g.id
	LEFT JOIN lessons l
	ON m.lesson_id = l.id
	INNER JOIN users u
	ON m.teacher_id = u.id
    LEFT JOIN excuses ex
    ON m.id = ex.mark_id
    LEFT JOIN users u2
    ON u2.id = ex.user_id
	WHERE m.journal_id = $1 and m.course = $2 and m.lesson_id is not NULL and m.user_id = $3
	ORDER BY updated_at ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID, course, userID)
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

func (m MarkModel) InsertMarks(tx *sql.Tx, marks []*model.Marks) error {
	stmt := table.Marks.INSERT(table.Marks.MutableColumns).
		MODELS(marks).
		ON_CONFLICT(table.Marks.UserID, table.Marks.LessonID, table.Marks.Type).
		WHERE(table.Marks.Type.IN(postgres.String(MarkAbsent), postgres.String(MarkLate), postgres.String(MarkNotDone))).
		DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (m MarkModel) UpdateMarks(tx *sql.Tx, marks []*model.Marks) error {
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
			tx.Rollback()
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
		tx.Rollback()
		return err
	}

	return nil
}

func (m MarkModel) DeleteMarksByStudentIDType(tx *sql.Tx, l []MarkByStudentIDType) error {
	var or []postgres.BoolExpression
	for _, m := range l {
		or = append(or, table.Marks.UserID.EQ(helpers.PostgresInt(m.StudentID)).AND(table.Marks.Type.EQ(postgres.String(m.Type))))
	}

	stmt := table.Marks.DELETE().
		WHERE(postgres.OR(or...))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (m MarkModel) GetStudentsForLesson(lessonID int) ([]*LessonStudent, error) {
	query := postgres.SELECT(
		table.Users.ID,
		table.Users.Name,
		table.Marks.ID,
		postgres.CASE().WHEN(table.Marks.Type.EQ(postgres.String(MarkLessonGrade))).THEN(postgres.String("grade")).ELSE(table.Marks.Type).AS("marks.type"),
		table.Marks.Comment,
		table.Grades.Identifier,
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
		marks := make([]*MinimalMark, 0, len(s.Lesson.Marks))
		for _, m := range s.Lesson.Marks {
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
		s.Lesson.Marks = marks
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

func (m MarkModel) InsertMark(mark *Mark) error {
	stmt := `INSERT INTO marks
	(user_id, lesson_id, course, journal_id, grade_id, comment, type, by, created_at, updated_at)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, mark.UserID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.Comment, mark.Type, mark.Teacher.ID, mark.CreatedAt, mark.UpdatedAt).Scan(&mark.ID)

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

	_, err := m.DB.ExecContext(ctx, stmt, mark.UserID, mark.Lesson.ID, mark.Course, mark.JournalID, mark.Grade.ID,
		mark.Comment, mark.Type, mark.Teacher.ID, mark.UpdatedAt, mark.ID)

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

	_, err := m.DB.ExecContext(ctx, stmt, markID)
	if err != nil {
		return err
	}

	return nil
}
