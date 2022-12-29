package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

const (
	RoleAdministrator = "admin"
	RoleTeacher       = "teacher"
	RoleParent        = "parent"
	RoleStudent       = "student"
)

var (
	ErrEmailAlreadyExists  = errors.New("an user with specified email already exists")
	ErrIDCodeAlreadyExists = errors.New("an user with specified ID code already exists")
	ErrNoSuchUser          = errors.New("no such user")
	ErrNoSuchUsers         = errors.New("no such users")
	ErrNoSuchStudents      = errors.New("no such students")
	ErrNotAStudent         = errors.New("not a student")
	ErrNoSuchParentForUser = errors.New("no such parent set for child")
	ErrNotAParent          = errors.New("not a parent")
)

var EmailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

type User = model.Users

type Student struct {
	Class   *ClassExt `json:"class,omitempty"`
	Parents []*User   `json:"parents,omitempty" alias:"parents"`
}

type UserExt struct {
	User
	Student   *Student `json:"student,omitempty"`
	SessionID *int     `json:"-" alias:"sessions.id"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserModel struct {
	DB *sql.DB
}

// DATABASE

func (m UserModel) AllUsers(archived bool) ([]*UserExt, error) {
	query := postgres.SELECT(table.Users.AllColumns.Except(table.Users.Password), table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears,
				table.ClassesYears.ClassID.EQ(table.Classes.ID).
					AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.Users.Archived.EQ(postgres.Bool(archived))).
		ORDER_BY(table.Users.ID.ASC())

	var users []*UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) SearchUser(name string) ([]*UserExt, error) {
	// todo: LIKE + LOWER -> ILIKE

	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role, table.Users.ClassID, table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(postgres.LOWER(table.Users.Name).LIKE(postgres.LOWER(postgres.String("%" + name + "%"))).
			AND(table.Users.Archived.IS_FALSE())).
		ORDER_BY(table.Users.Name.ASC())

	var users []*UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetUserByID(userID int) (*UserExt, error) {
	query := postgres.SELECT(table.Users.AllColumns, table.Classes.ID, table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
				AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.Users.ID.EQ(helpers.PostgresInt(userID)))

	var user UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m UserModel) GetUsersByRole(role string) ([]*UserExt, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users).
		WHERE(table.Users.Role.EQ(postgres.String(role))).
		ORDER_BY(table.Users.ID.ASC())

	var users []*UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) InsertUser(u *User) error {
	stmt := table.Users.INSERT(table.Users.MutableColumns.
		Except(table.Users.CreatedAt, table.Users.Active, table.Users.Archived)).
		MODEL(u).
		RETURNING(table.Users.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, u)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if strings.Contains(pgErr.Message, "email") {
				return ErrEmailAlreadyExists
			} else if strings.Contains(pgErr.Message, "id_code") {
				return ErrIDCodeAlreadyExists
			}
			return err
		} else {
			return err
		}
	}

	return nil
}

func (m UserModel) UpdateUser(u *UserExt) error {
	stmt := table.Users.UPDATE(table.Users.MutableColumns).
		MODEL(u).
		WHERE(table.Users.ID.EQ(helpers.PostgresInt(u.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if strings.Contains(pgErr.Message, "email") {
				return ErrEmailAlreadyExists
			} else if strings.Contains(pgErr.Message, "id_code") {
				return ErrIDCodeAlreadyExists
			}
			return err
		} else {
			return err
		}
	}

	return nil

}

func (m UserModel) GetAllUserIDs() ([]int, error) {
	query := postgres.SELECT(table.Users.ID).
		FROM(table.Users).
		WHERE(table.Users.Archived.IS_FALSE())

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m UserModel) GetAllStudentIDs() ([]int, error) {
	query := postgres.SELECT(table.Users.ID).
		FROM(table.Users).
		WHERE(table.Users.Role.EQ(postgres.String(RoleStudent)).AND(table.Users.Archived.IS_FALSE()))

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m UserModel) GetUserBySessionToken(plaintextToken string) (*UserExt, error) {
	hash := sha256.Sum256([]byte(plaintextToken))

	query := postgres.SELECT(table.Users.AllColumns, table.Classes.Name, table.Sessions.ID).
		FROM(table.Users.
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			INNER_JOIN(table.Sessions, table.Sessions.UserID.EQ(table.Users.ID))).
		WHERE(postgres.AND(
			table.Users.Archived.IS_FALSE(),
			table.Users.Active.IS_TRUE(),
			table.Sessions.Token.EQ(postgres.Bytea(hash[:])),
			table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC()))))

	var user UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &user)

	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrInvalidToken
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetUserByEmail(email string) (*UserExt, error) {
	query := postgres.SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(postgres.AND(
			table.Users.Email.EQ(postgres.String(email)),
			table.Users.Archived.IS_FALSE(),
			table.Users.Active.IS_TRUE(),
		))

	var user UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &user)

	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) AddParentToChild(parentID, childID int) error {
	stmt := table.ParentsChildren.INSERT(table.ParentsChildren.AllColumns).
		MODEL(model.ParentsChildren{
			ParentID: &parentID,
			ChildID:  &childID,
		}).
		ON_CONFLICT(table.ParentsChildren.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) RemoveParentFromChild(parentID, childID int) error {
	stmt := table.ParentsChildren.DELETE().
		WHERE(table.ParentsChildren.ParentID.EQ(helpers.PostgresInt(parentID)).
			AND(table.ParentsChildren.ChildID.EQ(helpers.PostgresInt(childID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) GetStudentByID(userID int) (*UserExt, error) {
	parent := table.Users.AS("parents")

	query := postgres.SELECT(
		table.Users.ID, table.Users.Name, table.Users.Email, table.Users.PhoneNumber, table.Users.IDCode, table.Users.BirthDate, table.Users.Role, table.Users.ClassID,
		table.Classes.ID, table.Classes.Name,
		parent.ID, parent.Name, parent.Email, parent.PhoneNumber, parent.IDCode, parent.BirthDate, parent.Role,
	).FROM(table.Users.
		INNER_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
		LEFT_JOIN(table.ParentsChildren, table.ParentsChildren.ChildID.EQ(table.Users.ID)).
		LEFT_JOIN(parent, parent.ID.EQ(table.ParentsChildren.ParentID))).
		WHERE(table.Users.ID.EQ(helpers.PostgresInt(userID)).
			AND(table.Users.Role.EQ(postgres.String(RoleStudent))))

	var user UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &user)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetParentsForChild(childID int) ([]*UserExt, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Email, table.Users.PhoneNumber, table.Users.IDCode, table.Users.BirthDate, table.Users.Role).
		FROM(table.Users.
			INNER_JOIN(table.ParentsChildren, table.ParentsChildren.ParentID.EQ(table.Users.ID))).
		WHERE(table.ParentsChildren.ChildID.EQ(helpers.PostgresInt(childID)))

	var users []*UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetChildrenForParent(parentID int) ([]*UserExt, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users.
			INNER_JOIN(table.ParentsChildren, table.ParentsChildren.ChildID.EQ(table.Users.ID))).
		WHERE(table.ParentsChildren.ParentID.EQ(helpers.PostgresInt(parentID)))

	var users []*UserExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) IsUserTeacherOrParentOfStudent(studentID, userID int) (bool, error) {
	query := postgres.SELECT(postgres.COUNT(postgres.Int32(1))).
		FROM(table.Users.
			LEFT_JOIN(table.ParentsChildren, table.ParentsChildren.ChildID.EQ(table.Users.ID)).
			LEFT_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Users.ClassID))).
		WHERE(table.Users.ID.EQ(helpers.PostgresInt(studentID)).
			AND(table.ParentsChildren.ParentID.EQ(helpers.PostgresInt(userID)).
				OR(table.TeachersClasses.TeacherID.EQ(helpers.PostgresInt(userID)))))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

func (m UserModel) IsUserTeacherOfStudent(studentID, userID int) (bool, error) {
	query := postgres.SELECT(postgres.COUNT(postgres.Int32(1))).
		FROM(table.Users.
			INNER_JOIN(table.TeachersClasses, table.TeachersClasses.ClassID.EQ(table.Users.ClassID))).
		WHERE(table.Users.ID.EQ(helpers.PostgresInt(studentID)).
			AND(table.TeachersClasses.TeacherID.EQ(helpers.PostgresInt(userID))))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

func (m UserModel) IsUserTeacherOfClass(userID, classID int) (bool, error) {
	query := postgres.SELECT(postgres.COUNT(postgres.Int32(1))).
		FROM(table.TeachersClasses).
		WHERE(table.TeachersClasses.TeacherID.EQ(helpers.PostgresInt(userID)).
			AND(table.TeachersClasses.ClassID.EQ(helpers.PostgresInt(classID))))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

func (m UserModel) IsUserParentOfStudent(studentID, userID int) (bool, error) {
	query := postgres.SELECT(postgres.COUNT(postgres.Int32(1))).
		FROM(table.ParentsChildren).
		WHERE(table.ParentsChildren.ChildID.EQ(helpers.PostgresInt(studentID)).
			AND(table.ParentsChildren.ParentID.EQ(helpers.PostgresInt(userID))))

	var result []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &result)
	if err != nil {
		return false, err
	}

	return result[0] > 0, nil
}

func (m UserModel) ArchiveUsersByClassID(classID int) error {
	stmt := table.Users.UPDATE(table.Users.Archived).
		SET(postgres.Bool(true)).
		WHERE(table.Users.ClassID.EQ(helpers.PostgresInt(classID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
