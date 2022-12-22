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
	Course   *int            `json:"course,omitempty" alias:"coursenr"`
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
	stmt := table.Journals.DELETE().
		WHERE(table.Journals.ID.EQ(helpers.PostgresInt(journalID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
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

func (m JournalModel) InsertStudentsIntoJournal(studentIDs []int, journalID int) error {
	var sjs []model.StudentsJournals
	for _, sid := range studentIDs {
		sid := sid
		sjs = append(sjs, model.StudentsJournals{
			StudentID: &sid,
			JournalID: &journalID,
		})
	}

	stmt := table.StudentsJournals.INSERT(table.StudentsJournals.AllColumns).
		MODELS(sjs).
		ON_CONFLICT(table.StudentsJournals.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m JournalModel) DeleteStudentFromJournal(studentID, journalID int) error {
	stmt := table.StudentsJournals.DELETE().
		WHERE(table.StudentsJournals.StudentID.EQ(helpers.PostgresInt(studentID)).
			AND(table.StudentsJournals.JournalID.EQ(helpers.PostgresInt(journalID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := stmt.ExecContext(ctx, m.DB)
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

func (m JournalModel) GetStudentsByJournalID(journalID int) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role, table.Classes.ID, table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			INNER_JOIN(table.StudentsJournals, table.StudentsJournals.StudentID.EQ(table.Users.ID)).
			INNER_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.StudentsJournals.JournalID.EQ(helpers.PostgresInt(journalID))).
		ORDER_BY(table.Users.Name.ASC())

	var students []*NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &students)
	if err != nil {
		return nil, err
	}

	return students, nil
}

func (m JournalModel) GetJournalsByStudent(studentID, yearID int) ([]*NJournal, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(
		table.Journals.ID,
		teacher.ID,
		teacher.Name,
		table.Subjects.ID,
		table.Subjects.Name,
	).FROM(table.Journals.
		INNER_JOIN(table.StudentsJournals, table.StudentsJournals.JournalID.EQ(table.Journals.ID)).
		INNER_JOIN(table.Years, table.Years.ID.EQ(table.Journals.YearID)).
		LEFT_JOIN(table.TeachersJournals, table.TeachersJournals.JournalID.EQ(table.Journals.ID)).
		LEFT_JOIN(teacher, teacher.ID.EQ(table.TeachersJournals.TeacherID)).
		INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID))).
		WHERE(table.StudentsJournals.StudentID.EQ(helpers.PostgresInt(studentID)).
			AND(table.Years.ID.EQ(helpers.PostgresInt(yearID)))).
		ORDER_BY(table.Subjects.Name.ASC())

	var journals []*NJournal

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &journals)
	if err != nil {
		return nil, err
	}

	return journals, nil
}

func (m JournalModel) IsUserInJournal(studentID, journalID int) (bool, error) {
	query := postgres.SELECT(postgres.COUNT(postgres.Int32(1))).
		FROM(table.StudentsJournals).
		WHERE(table.StudentsJournals.StudentID.EQ(helpers.PostgresInt(studentID)).
			AND(table.StudentsJournals.JournalID.EQ(helpers.PostgresInt(journalID))))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

// todo: not needed?
// func (m JournalModel) DoesParentHaveChildInJournal(parentID, journalID int) (bool, error) {
// 	query := postgres.SELECT(postgres.COUNT(postgres.Int32(1))).
// 		FROM(table.ParentsChildren.
// 			INNER_JOIN(table.StudentsJournals, table.StudentsJournals.StudentID.EQ(table.ParentsChildren.ChildID))).
// 		WHERE(table.ParentsChildren.ParentID.EQ(helpers.PostgresInt(parentID)).
// 			AND(table.StudentsJournals.JournalID.EQ(helpers.PostgresInt(journalID))))

// 	var result []int

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	err := query.QueryContext(ctx, m.DB, &result)
// 	if err != nil {
// 		return false, err
// 	}

// 	return result[0] > 0, nil
// }

func (m JournalModel) SetJournalLastUpdated(journalID int) error {
	stmt := table.Journals.UPDATE(table.Journals.LastUpdated).
		SET(time.Now().UTC()).
		WHERE(table.Journals.ID.EQ(helpers.PostgresInt(journalID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
