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
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-jet/jet/v2/postgres"
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
	Class *NClass `json:"class,omitempty"`
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

func (m UserModel) InsertUser(u *User) error {
	stmt := `INSERT INTO users
	(name, email, phone_number, id_code, birth_date, password, role, class_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt,
		u.Name,
		u.Email,
		u.PhoneNumber,
		u.IdCode,
		u.BirthDate.Time,
		u.Password.Hashed,
		u.Role,
		u.Class.ID,
		u.CreatedAt).Scan(&u.ID)

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
	stmt := `UPDATE users SET (name, email, phone_number, id_code, birth_date, password, role, class_id, created_at, active, archived) =
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	WHERE id = $12`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt,
		u.Name,
		u.Email,
		u.PhoneNumber,
		u.IDCode,
		u.BirthDate.Time,
		u.Password.Hashed,
		u.Role,
		u.Class.ID,
		u.CreatedAt,
		u.Active,
		u.Archived,
		u.ID)

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
	query := `SELECT
	array(SELECT id	FROM users WHERE archived is FALSE)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRowContext(ctx, query).Scan(pgtype.NewMap().SQLScanner(&ids))
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m UserModel) GetAllStudentIDs() ([]int, error) {
	query := `SELECT
	array(SELECT id	FROM users WHERE role = 'student' AND archived is FALSE)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRowContext(ctx, query).Scan(pgtype.NewMap().SQLScanner(&ids))
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m UserModel) GetUserBySessionToken(plaintextToken string) (*User, error) {
	hash := sha256.Sum256([]byte(plaintextToken))

	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.password, u.role, u.class_id, c.name, u.created_at, u.active, u.archived, s.id
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	INNER JOIN sessions s
	ON u.id = s.user_id
	WHERE u.archived is FALSE
	AND u.active is TRUE
	AND s.token_hash = $1
	AND s.expires > $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(types.Date)
	user.Class = new(Class)

	err := m.DB.QueryRowContext(ctx, query, hash[:], time.Now().UTC()).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Password.Hashed,
		&user.Role,
		&user.Class.ID,
		&user.Class.Name,
		&user.CreatedAt,
		&user.Active,
		&user.Archived,
		&user.SessionID,
	)

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

func (m UserModel) GetUserByEmail(email string) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.password, u.role, u.created_at, u.active, u.archived
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	WHERE u.email = $1 AND u.archived is FALSE AND u.active is TRUE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(types.Date)

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Password.Hashed,
		&user.Role,
		&user.CreatedAt,
		&user.Active,
		&user.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) AddParentToChild(parentID, childID int) error {
	stmt := `INSERT INTO parents_children
	(parent_id, child_id)
	VALUES
	($1, $2)
	ON CONFLICT DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, parentID, childID)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) RemoveParentFromChild(parentID, childID int) error {
	stmt := `DELETE FROM parents_children
	WHERE parent_id = $1 and child_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, parentID, childID)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) GetStudentByID(userID int) (*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.role, u.class_id, c.name, u2.id, u2.name
	FROM users u
	LEFT JOIN classes c
	ON u.class_id = c.id
	LEFT JOIN users u2
	ON c.teacher_id = u2.id
	WHERE u.id = $1 and u.role = 'student'`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	user.BirthDate = new(types.Date)
	user.Class = &Class{Teacher: new(User)}

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.IdCode,
		&user.BirthDate.Time,
		&user.Role,
		&user.Class.ID,
		&user.Class.Name,
		&user.Class.Teacher.ID,
		&user.Class.Teacher.Name,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetParentsForChild(childID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.email, u.phone_number, u.id_code, u.birth_date, u.role
	FROM users u
	INNER JOIN parents_children pc
	ON u.id = pc.parent_id
	WHERE pc.child_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, childID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		user.BirthDate = new(types.Date)

		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PhoneNumber,
			&user.IdCode,
			&user.BirthDate.Time,
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

func (m UserModel) GetChildrenForParent(parentID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.role
	FROM users u
	INNER JOIN parents_children pc
	ON u.id = pc.child_id
	WHERE pc.parent_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, parentID)
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

func (m UserModel) IsUserTeacherOrParentOfStudent(studentID, userID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM users s
	LEFT JOIN parents_children pc
	ON s.id = pc.child_id
	LEFT JOIN classes c
	ON s.class_id = c.id
	WHERE s.id = $1 AND (pc.parent_id = $2 OR (c.teacher_id = $2 AND c.archived is FALSE))`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRowContext(ctx, query, studentID, userID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (m UserModel) IsUserParentOfStudent(studentID, userID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM parents_children
	WHERE child_id = $1 AND parent_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRowContext(ctx, query, studentID, userID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (m UserModel) ArchiveUsersByClassID(classID int) error {
	stmt := `UPDATE users u
	SET archived = true
	WHERE u.class_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, classID)
	if err != nil {
		return err
	}

	return nil
}
