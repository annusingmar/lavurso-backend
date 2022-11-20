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
	"golang.org/x/crypto/bcrypt"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/types"
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

type User struct {
	ID          *int           `json:"id"`
	Name        *string        `json:"name,omitempty"`
	Email       *string        `json:"email,omitempty"`
	PhoneNumber *string        `json:"phone_number,omitempty"`
	IdCode      *int64         `json:"id_code,omitempty"`
	BirthDate   *types.Date    `json:"birth_date,omitempty"`
	Password    types.Password `json:"-"`
	Role        *string        `json:"role,omitempty"`
	CreatedAt   *time.Time     `json:"created_at,omitempty"`
	Active      *bool          `json:"active,omitempty"`
	Archived    *bool          `json:"archived,omitempty"`
	Class       *Class         `json:"class,omitempty"`
	Marks       []*Mark        `json:"marks,omitempty"`
	Children    []*User        `json:"children,omitempty"`
	SessionID   *int           `json:"-"`
}

type NUser struct {
	model.Users
	Class     *NClass `json:"class,omitempty"`
	SessionID *int    `json:"-" alias:"sessions.id"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) HashPassword(plaintext string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return nil, err
	}
	return hashed, nil
}

// DATABASE

func (m UserModel) AllUsers(archived bool) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.AllColumns.Except(table.Users.Password), table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears,
				table.ClassesYears.ClassID.EQ(table.Classes.ID).
					AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.Users.Archived.EQ(postgres.Bool(archived))).
		ORDER_BY(table.Users.ID.ASC())

	var users []*NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) SearchUser(name string) ([]*NUser, error) {
	// todo: LIKE + LOWER -> ILIKE

	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role, table.Users.ClassID, table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(postgres.LOWER(table.Users.Name).LIKE(postgres.LOWER(postgres.String("%" + name + "%"))).
			AND(table.Users.Archived.IS_FALSE())).
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

func (m UserModel) GetUserByID(userID int) (*NUser, error) {
	query := postgres.SELECT(table.Users.AllColumns, table.Classes.ID, table.Classes.Name, table.ClassesYears.DisplayName).
		FROM(table.Users.
			LEFT_JOIN(table.Years, table.Years.Current.IS_TRUE()).
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.ClassID.EQ(table.Classes.ID).
				AND(table.ClassesYears.YearID.EQ(table.Years.ID)))).
		WHERE(table.Users.ID.EQ(postgres.Int32(int32(userID))))

	var user NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m UserModel) GetUsersByRole(role string) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users).
		WHERE(table.Users.Role.EQ(postgres.String(role))).
		ORDER_BY(table.Users.ID.ASC())

	var users []*NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) InsertUser(u *NUser) error {
	stmt := table.Users.INSERT(
		table.Users.Name,
		table.Users.Email,
		table.Users.PhoneNumber,
		table.Users.IDCode,
		table.Users.BirthDate,
		table.Users.Password,
		table.Users.Role,
		table.Users.ClassID,
	).MODEL(u).RETURNING(table.Users.ID)

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

func (m UserModel) UpdateUser(u *NUser) error {
	stmt := table.Users.UPDATE(table.Users.MutableColumns).
		MODEL(u).
		WHERE(table.Users.ID.EQ(postgres.Int32(int32(u.ID))))

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

func (m UserModel) GetUserBySessionToken(plaintextToken string) (*NUser, error) {
	hash := sha256.Sum256([]byte(plaintextToken))

	query := postgres.SELECT(table.Users.AllColumns, table.Classes.Name, table.Sessions.ID).
		FROM(table.Users.
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
			INNER_JOIN(table.Sessions, table.Sessions.UserID.EQ(table.Users.ID))).
		WHERE(postgres.AND(
			table.Users.Archived.IS_FALSE(),
			table.Users.Active.IS_TRUE(),
			table.Sessions.TokenHash.EQ(postgres.Bytea(hash[:])),
			table.Sessions.Expires.GT(postgres.TimestampzT(time.Now().UTC()))))

	var user NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &user)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrInvalidToken
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetUserByEmail(email string) (*NUser, error) {
	query := postgres.SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(postgres.AND(
			table.Users.Email.EQ(postgres.String(email)),
			table.Users.Archived.IS_FALSE(),
			table.Users.Active.IS_TRUE(),
		))

	var user NUser

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
		WHERE(table.ParentsChildren.ParentID.EQ(postgres.Int32(int32(parentID))).
			AND(table.ParentsChildren.ChildID.EQ(postgres.Int32(int32(childID)))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) GetStudentByID(userID int) (*NUser, error) {
	teacher := table.Users.AS("teacher")

	query := postgres.SELECT(
		table.Users.ID, table.Users.Name, table.Users.Email, table.Users.PhoneNumber, table.Users.IDCode, table.Users.BirthDate, table.Users.Role, table.Users.ClassID,
		table.Classes.ID, table.Classes.Name,
		teacher.ID, teacher.Name,
	).FROM(table.Users.
		LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID)).
		LEFT_JOIN(teacher, teacher.ID.EQ(table.Classes.TeacherID))).
		WHERE(table.Users.ID.EQ(postgres.Int32(int32(userID))).
			AND(table.Users.Role.EQ(postgres.String(RoleStudent))))

	var user NUser

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

func (m UserModel) GetParentsForChild(childID int) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Email, table.Users.PhoneNumber, table.Users.IDCode, table.Users.BirthDate, table.Users.Role).
		FROM(table.Users.
			INNER_JOIN(table.ParentsChildren, table.ParentsChildren.ParentID.EQ(table.Users.ID))).
		WHERE(table.ParentsChildren.ChildID.EQ(postgres.Int32(int32(childID))))

	var users []*NUser

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetChildrenForParent(parentID int) ([]*NUser, error) {
	query := postgres.SELECT(table.Users.ID, table.Users.Name, table.Users.Role).
		FROM(table.Users.
			INNER_JOIN(table.ParentsChildren, table.ParentsChildren.ChildID.EQ(table.Users.ID))).
		WHERE(table.ParentsChildren.ParentID.EQ(postgres.Int32(int32(parentID))))

	var users []*NUser

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
			LEFT_JOIN(table.Classes, table.Classes.ID.EQ(table.Users.ClassID))).
		WHERE(table.Users.ID.EQ(postgres.Int32(int32(studentID))).
			AND(table.ParentsChildren.ParentID.EQ(postgres.Int32(int32(userID))).
				OR(table.Classes.TeacherID.EQ(postgres.Int32(int32(userID))))))

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
		WHERE(table.ParentsChildren.ChildID.EQ(postgres.Int32(int32(studentID))).
			AND(table.ParentsChildren.ParentID.EQ(postgres.Int32(int32(userID)))))

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
		WHERE(table.Users.ClassID.EQ(postgres.Int32(int32(classID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
