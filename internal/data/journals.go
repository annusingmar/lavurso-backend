package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNoSuchJournal    = errors.New("no such journal")
	ErrUserNotInJournal = errors.New("user not in journal")
)

type NJournal struct {
	model.Journals
	Subject  *model.Subjects `json:"subject,omitempty"`
	Teachers []*model.Users  `json:"teachers,omitempty" alias:"teachers"`
	Year     *model.Years    `json:"year,omitempty"`
	Course   int             `json:"course,omitempty" alias:"coursenr"`
}

type Journal struct {
	ID          int             `json:"id"`
	Name        *string         `json:"name,omitempty"`
	Teacher     *User           `json:"teacher,omitempty"`
	Subject     *Subject        `json:"subject,omitempty"`
	Year        *Year           `json:"year,omitempty"`
	LastUpdated *time.Time      `json:"last_updated,omitempty"`
	Courses     []int           `json:"courses,omitempty"`
	Marks       map[int][]*Mark `json:"marks,omitempty"`
}

type JournalModel struct {
	DB *sql.DB
}

func (j *NJournal) IsUserTeacherOfJournal(userID int) bool {
	for _, t := range j.Teachers {
		if t.ID == userID {
			return true
		}
	}
	return false
}

func (m JournalModel) AllJournals(yearID int) ([]*NJournal, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(
		table.Journals.AllColumns,
		table.Subjects.ID, table.Subjects.Name,
		teacher.ID, teacher.Name, teacher.Role,
		table.Years.ID, table.Years.DisplayName).
		FROM(table.Journals.
			LEFT_JOIN(table.TeachersJournals, table.TeachersJournals.JournalID.EQ(table.Journals.ID)).
			LEFT_JOIN(teacher, teacher.ID.EQ(table.TeachersJournals.TeacherID)).
			INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID)).
			INNER_JOIN(table.Years, table.Years.ID.EQ(table.Journals.YearID))).
		WHERE(table.Years.ID.EQ(helpers.PostgresInt(yearID))).
		ORDER_BY(table.Journals.LastUpdated.DESC())

	var journals []*NJournal

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &journals)
	if err != nil {
		return nil, err
	}

	return journals, nil
}

func (m JournalModel) GetJournalByID(journalID int) (*NJournal, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(
		table.Journals.AllColumns,
		table.Subjects.ID, table.Subjects.Name,
		teacher.ID, teacher.Name, teacher.Role,
		table.Years.ID, table.Years.DisplayName,
		postgres.SELECT(postgres.MAX(table.Lessons.Course)).
			FROM(table.Lessons).
			WHERE(table.Lessons.JournalID.EQ(table.Journals.ID)).AS("njournal.coursenr")).
		FROM(table.Journals.
			LEFT_JOIN(table.TeachersJournals, table.TeachersJournals.JournalID.EQ(table.Journals.ID)).
			LEFT_JOIN(teacher, teacher.ID.EQ(table.TeachersJournals.TeacherID)).
			INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID)).
			INNER_JOIN(table.Years, table.Years.ID.EQ(table.Journals.YearID))).
		WHERE(table.Journals.ID.EQ(helpers.PostgresInt(journalID)))

	var journal NJournal

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &journal)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchJournal
		default:
			return nil, err
		}
	}

	return &journal, nil
}

func (m JournalModel) InsertJournal(j *model.Journals, teacherID int) error {
	stmt := table.Journals.INSERT(table.Journals.Name, table.Journals.SubjectID, table.Journals.YearID).
		MODEL(j).
		RETURNING(table.Journals.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, j)
	if err != nil {
		return err
	}

	stmt = table.TeachersJournals.INSERT(table.TeachersJournals.AllColumns).
		MODEL(model.TeachersJournals{TeacherID: &teacherID, JournalID: &j.ID})

	_, err = stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m JournalModel) UpdateJournal(j *NJournal, teacherIDs []int) error {
	stmt := table.Journals.UPDATE(table.Journals.Name, table.Journals.LastUpdated).
		MODEL(j).
		WHERE(table.Journals.ID.EQ(helpers.PostgresInt(j.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	var tjs []model.TeachersJournals
	var tids []postgres.Expression
	for _, tid := range teacherIDs {
		tid := tid
		tjs = append(tjs, model.TeachersJournals{
			TeacherID: &tid,
			JournalID: &j.ID,
		})
		tids = append(tids, helpers.PostgresInt(tid))
	}

	var deletestmt postgres.DeleteStatement

	if tjs != nil {
		insertstmt := table.TeachersJournals.INSERT(table.TeachersJournals.AllColumns).
			MODELS(tjs).
			ON_CONFLICT(table.TeachersJournals.AllColumns...).DO_NOTHING()

		_, err := insertstmt.ExecContext(ctx, m.DB)
		if err != nil {
			return err
		}

		deletestmt = table.TeachersJournals.DELETE().WHERE(table.TeachersJournals.TeacherID.NOT_IN(tids...).
			AND(table.TeachersJournals.JournalID.EQ(helpers.PostgresInt(j.ID))))
	} else {
		deletestmt = table.TeachersJournals.DELETE().WHERE(table.TeachersJournals.JournalID.EQ(helpers.PostgresInt(j.ID)))
	}

	_, err = deletestmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m JournalModel) DeleteJournal(journalID int) error {
	stmt := `DELETE FROM journals
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, journalID)
	if err != nil {
		return err
	}
	return nil
}

func (m JournalModel) GetJournalsForTeacher(teacherID, yearID int) ([]*NJournal, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(
		table.Journals.AllColumns,
		table.Subjects.ID, table.Subjects.Name,
		teacher.ID, teacher.Name, teacher.Role,
		table.Years.ID, table.Years.DisplayName).
		FROM(table.Journals.
			INNER_JOIN(table.TeachersJournals, table.TeachersJournals.JournalID.EQ(table.Journals.ID)).
			INNER_JOIN(teacher, teacher.ID.EQ(table.TeachersJournals.TeacherID)).
			INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID)).
			INNER_JOIN(table.Years, table.Years.ID.EQ(table.Journals.YearID))).
		WHERE(table.Journals.ID.IN(
			postgres.SELECT(table.Journals.ID).
				FROM(table.Journals.
					INNER_JOIN(table.TeachersJournals, table.TeachersJournals.JournalID.EQ(table.Journals.ID))).
				WHERE(table.TeachersJournals.TeacherID.EQ(helpers.PostgresInt(teacherID))),
		).AND(table.Journals.YearID.EQ(helpers.PostgresInt(yearID)))).
		ORDER_BY(table.Journals.LastUpdated.DESC())

	var journals []*NJournal

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &journals)
	if err != nil {
		return nil, err
	}

	return journals, nil
}

func (m JournalModel) InsertStudentIntoJournal(studentID, journalID int) error {
	stmt := `INSERT INTO
	students_journals
	(student_id, journal_id)
	VALUES
	($1, $2)
	ON CONFLICT DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, studentID, journalID)
	if err != nil {
		return err
	}

	return nil
}

func (m JournalModel) DeleteStudentFromJournal(studentID, journalID int) error {
	stmt := `DELETE FROM
	students_journals
	WHERE student_id = $1 and journal_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, stmt, studentID, journalID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrUserNotInJournal
	}

	return nil
}

func (m JournalModel) GetStudentsByJournalID(journalID int) ([]*User, error) {
	query := `SELECT id, name, role
	FROM users u
	INNER JOIN students_journals uj
	ON uj.student_id = u.id
	WHERE uj.journal_id = $1
	ORDER BY name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m JournalModel) GetJournalsByStudent(userID, yearID int) ([]*Journal, error) {
	query := `SELECT j.id, j.teacher_id, u.name, u.role, j.subject_id, s.name, y.id, y.display_name, j.last_updated, array(SELECT DISTINCT course FROM lessons WHERE journal_id = j.id)
	FROM journals j
	INNER JOIN users u
	ON j.teacher_id = u.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	INNER JOIN students_journals uj
	ON uj.journal_id = j.id
	INNER JOIN years y
	ON j.year_id = y.id
	WHERE uj.student_id = $1 AND y.id = $2
	ORDER BY s.name ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID, yearID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var journals []*Journal

	for rows.Next() {
		var journal Journal
		journal.Teacher = new(User)
		journal.Subject = new(Subject)
		journal.Year = new(Year)

		err = rows.Scan(
			&journal.ID,
			&journal.Teacher.ID,
			&journal.Teacher.Name,
			&journal.Teacher.Role,
			&journal.Subject.ID,
			&journal.Subject.Name,
			&journal.Year.ID,
			&journal.Year.DisplayName,
			&journal.LastUpdated,
			pgtype.NewMap().SQLScanner(&journal.Courses),
		)
		if err != nil {
			return nil, err
		}

		journals = append(journals, &journal)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return journals, nil
}

func (m JournalModel) IsUserInJournal(userID, journalID int) (bool, error) {
	query := `SELECT COUNT(1) FROM students_journals
	WHERE student_id = $1 and journal_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRowContext(ctx, query, userID, journalID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (m JournalModel) DoesParentHaveChildInJournal(parentID, journalID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM parents_children pc
	INNER JOIN students_journals uj
	ON pc.child_id = uj.student_id
	WHERE pc.parent_id = $1
	AND uj.journal_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRowContext(ctx, query, parentID, journalID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (m JournalModel) SetJournalLastUpdated(journalID int) error {
	stmt := `UPDATE journals
	SET last_updated = $1
	WHERE id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, time.Now().UTC(), journalID)
	if err != nil {
		return err
	}

	return nil
}
