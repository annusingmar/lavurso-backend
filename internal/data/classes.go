package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

var (
	ErrNoClassForUser = errors.New("no class set for user")
	ErrNoSuchClass    = errors.New("no such class")
	ErrClassArchived  = errors.New("class is archived")
)

type Class struct {
	ID          *int    `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Teacher     *User   `json:"teacher,omitempty"`
}

type NClass struct {
	model.Classes
	DisplayName *string      `json:"display_name,omitempty" alias:"classes_years.display_name"`
	Teacher     *model.Users `json:"teacher,omitempty" alias:"teacher"`
}

type ClassModel struct {
	DB *sql.DB
}

// DATABASE

func (m ClassModel) InsertClass(c *model.Classes) error {
	stmt := table.Classes.INSERT(table.Classes.MutableColumns).
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

func (m ClassModel) AllClasses(current bool) ([]*NClass, error) {
	teacher := table.Users.AS("teacher")

	query := postgres.SELECT(table.Classes.AllColumns, table.ClassesYears.DisplayName, teacher.ID, teacher.Name, teacher.Role)

	if current {
		query = query.
			FROM(table.Classes.
				LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
				INNER_JOIN(teacher, teacher.ID.EQ(table.Classes.TeacherID)).
				INNER_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
					AND(table.ClassesYears.YearID.EQ(table.Years.ID))))
	} else {
		query = query.
			FROM(table.Classes.
				LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
				INNER_JOIN(teacher, teacher.ID.EQ(table.Classes.TeacherID)).
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

func (m ClassModel) UpdateClass(c *NClass) error {
	stmt := table.Classes.UPDATE(table.Classes.MutableColumns).
		MODEL(c).
		WHERE(table.Classes.ID.EQ(postgres.Int32(int32(c.ID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil

}

func (m ClassModel) GetClassByID(classID int) (*NClass, error) {
	teacher := table.Users.AS("teacher")

	query := postgres.SELECT(table.Classes.AllColumns, teacher.ID, teacher.Name, teacher.Role).
		FROM(table.Classes.
			INNER_JOIN(teacher, teacher.ID.EQ(table.Classes.TeacherID))).
		WHERE(table.Classes.ID.EQ(postgres.Int32(int32(classID))))

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
	teacher := table.Users.AS("teacher")

	query := postgres.SELECT(table.Classes.AllColumns, table.ClassesYears.DisplayName, teacher.ID, teacher.Name, teacher.Role).
		FROM(table.Classes.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			INNER_JOIN(teacher, teacher.ID.EQ(table.Classes.TeacherID)).
			INNER_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
				AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.Classes.TeacherID.EQ(postgres.Int32(int32(teacherID))))

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
		WHERE(table.Users.ClassID.EQ(postgres.Int32(int32(classID)))).
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
