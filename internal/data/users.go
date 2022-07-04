package data

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/validator"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdministrator = "admin"
	RoleParent        = "parent"
	RoleStudent       = "student"
)

var (
	ErrEmailAlreadyExists = errors.New("an user with specified email already exists")
	ErrNoSuchUser         = errors.New("no such user")
	ErrNoSuchUsers        = errors.New("no such users")
	ErrEditConflict       = errors.New("edit conflict, try again")
	ErrNotAStudent        = errors.New("not a student")
)

var EmailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
	Version   int       `json:"version"`
}

type Password struct {
	Hashed    []byte
	Plaintext string
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m UserModel) HashPassword(plaintext string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return nil, err
	}
	return hashed, err
}

func (m UserModel) ValidateUser(v *validator.Validator, u *User) {
	v.Check(u.Name != "", "name", "must be provided")
	v.Check(u.Email != "", "email", "must be provided")
	v.Check(EmailRegex.MatchString(u.Email), "email", "must be a valid email address")
	v.Check(u.Role == RoleAdministrator || u.Role == RoleParent || u.Role == RoleStudent, "role", "must be valid role")
}

func (m UserModel) ValidatePassword(v *validator.Validator, u *User) {
	v.Check(u.Password.Plaintext != "", "password", "must be provided")
}

// DATABASE

func (m UserModel) AllUsers() ([]*User, error) {
	query := `SELECT id, name, email, password, role, created_at, active, version
	FROM users
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
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
			&user.Email,
			&user.Password.Hashed,
			&user.Role,
			&user.CreatedAt,
			&user.Active,
			&user.Version,
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

func (m UserModel) GetUserByID(userID int) (*User, error) {
	query := `SELECT id, name, email, password, role, created_at, active, version
	FROM users
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := m.DB.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.Hashed,
		&user.Role,
		&user.CreatedAt,
		&user.Active,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) InsertUser(u *User) error {
	stmt := `INSERT INTO users
	(name, email, password, role, created_at, active, version)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, u.Name, u.Email, u.Password.Hashed, u.Role, u.CreatedAt, u.Active, u.Version).Scan(&u.ID)
	if err != nil {
		switch {
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`:
			return ErrEmailAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) UpdateUser(u *User) error {
	stmt := `UPDATE users SET (name, email, password, role, active, version) =
	($1, $2, $3, $4, $5, version+1)
	WHERE id = $6
	RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, u.Name, u.Email, u.Password.Hashed, u.Role, u.Active, u.ID).Scan(&u.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil

}

func (m UserModel) GetAllUserIDs() ([]int, error) {
	query := `SELECT
	array(SELECT id	FROM users)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRow(ctx, query).Scan(&ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
