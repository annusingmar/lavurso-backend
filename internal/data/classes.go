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
	ErrNoClassForUser = errors.New("no class set for user")
	ErrNoSuchClass    = errors.New("no such class")
	ErrClassArchived  = errors.New("class is archived")
)

type NClass struct {
	model.Classes
	DisplayName *string        `json:"display_name,omitempty" alias:"classes_years.display_name"`
	Teachers    []*model.Users `json:"teachers,omitempty" alias:"teachers"`
}

type ClassModel struct {
	DB *sql.DB
}

// DATABASE

func (m ClassModel) InsertClass(c *model.Classes) error {
	stmt := table.Classes.INSERT(table.Classes.Name).
		MODEL(c).
		RETURNING(table.Classes.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, c)
	if err != nil {
		return err
	}

	return nil
}

func (m ClassModel) UpdateClass(c *NClass) error {
	stmt := table.Classes.UPDATE(table.Classes.Name).
		MODEL(c).
		WHERE(table.Classes.ID.EQ(helpers.PostgresInt(c.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m ClassModel) SetClassTeachers(classID int, teacherIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var tcs []model.TeachersClasses
	var tids []postgres.Expression
	for _, tid := range teacherIDs {
		tid := tid
		tcs = append(tcs, model.TeachersClasses{
			TeacherID: &tid,
			ClassID:   &classID,
		})
		tids = append(tids, helpers.PostgresInt(tid))
	}

	var deletestmt postgres.DeleteStatement

	if tcs != nil {
		insertstmt := table.TeachersClasses.INSERT(table.TeachersClasses.AllColumns).
			MODELS(tcs).
			ON_CONFLICT(table.TeachersClasses.AllColumns...).DO_NOTHING()

		_, err := insertstmt.ExecContext(ctx, m.DB)
		if err != nil {
			return err
		}

		deletestmt = table.TeachersClasses.DELETE().WHERE(table.TeachersClasses.TeacherID.NOT_IN(tids...).
			AND(table.TeachersClasses.ClassID.EQ(helpers.PostgresInt(classID))))
	} else {
		deletestmt = table.TeachersClasses.DELETE().WHERE(table.TeachersClasses.ClassID.EQ(helpers.PostgresInt(classID)))
	}

	_, err := deletestmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m ClassModel) AllClasses(current bool) ([]*NClass, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(table.Classes.AllColumns, table.ClassesYears.DisplayName, teacher.ID, teacher.Name, teacher.Role)

	if current {
		query = query.
			FROM(table.Classes.
				LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
				LEFT_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Classes.ID)).
				LEFT_JOIN(teacher, teacher.ID.EQ(table.TeachersClasses.TeacherID)).
				INNER_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
					AND(table.ClassesYears.YearID.EQ(table.Years.ID))))
	} else {
		query = query.
			FROM(table.Classes.
				LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
				LEFT_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Classes.ID)).
				LEFT_JOIN(teacher, teacher.ID.EQ(table.TeachersClasses.TeacherID)).
				LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
					AND(table.ClassesYears.YearID.EQ(table.Years.ID))))
	}

	var classes []*NClass

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &classes)
	if err != nil {
		return nil, err
	}

	return classes, nil
}

func (m ClassModel) GetAllClassIDs() ([]int, error) {
	query := postgres.SELECT(table.Classes.ID).
		FROM(table.Classes)

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m ClassModel) GetClassByID(classID int) (*NClass, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(table.Classes.AllColumns, teacher.ID, teacher.Name, teacher.Role).
		FROM(table.Classes.
			LEFT_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Classes.ID)).
			LEFT_JOIN(teacher, teacher.ID.EQ(table.TeachersClasses.TeacherID))).
		WHERE(table.Classes.ID.EQ(helpers.PostgresInt(classID)))

	var class NClass

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &class)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchClass
		default:
			return nil, err
		}
	}

	return &class, nil
}

func (m ClassModel) GetCurrentYearClassesForTeacher(teacherID int) ([]*NClass, error) {
	teacher := table.Users.AS("teachers")

	query := postgres.SELECT(table.Classes.AllColumns, table.ClassesYears.DisplayName, teacher.ID, teacher.Name, teacher.Role).
		FROM(table.Classes.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			INNER_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Classes.ID)).
			INNER_JOIN(teacher, teacher.ID.EQ(table.TeachersClasses.TeacherID)).
			INNER_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
				AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.Classes.ID.IN(
			postgres.SELECT(table.Classes.ID).FROM(table.Classes.
				INNER_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Classes.ID))).
				WHERE(table.TeachersClasses.TeacherID.EQ(helpers.PostgresInt(teacherID))),
		))

	var classes []*NClass

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &classes)
	if err != nil {
		return nil, err
	}

	return classes, nil
}

func (m ClassModel) GetUsersForClassID(classID int) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users).
		WHERE(table.Users.ClassID.EQ(helpers.PostgresInt(classID))).
		ORDER_BY(table.Users.Name.ASC())

	var users []*NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}
