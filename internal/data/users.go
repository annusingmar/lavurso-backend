package data

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	Administrator = iota + 1
	Parent
	Student
)

var (
	ErrEmailAlreadyExists = errors.New("an user with specified email already exists")
	ErrNoSuchUser         = errors.New("no such user")
	ErrEditConflict       = errors.New("edit conflict, try again")
)

var EmailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Role      int       `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type Password struct {
	Hashed    []byte
	Plaintext string
}

type UserModel struct {
	DB *sql.DB
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
	v.Check(u.Role > 0 && u.Role < 4, "role", "must be {1,2,3}")
}

func (m UserModel) ValidatePassword(v *validator.Validator, u *User) {
	v.Check(u.Password.Plaintext != "", "password", "must be provided")
}

// DATABASE

func (m UserModel) AllUsers() ([]*User, error) {
	query := `SELECT id, name, email, password, role, created_at, version FROM users`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
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

func (m UserModel) GetUserById(userID int) (*User, error) {
	query := `SELECT id, name, email, password, role, created_at, version
	FROM users
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.Hashed,
		&user.Role,
		&user.CreatedAt,
		&user.Version,
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

func (m UserModel) InsertUser(u *User) error {
	stmt := `INSERT INTO users
	(name, email, password, role, created_at, version)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, u.Name, u.Email, u.Password.Hashed, u.Role, u.CreatedAt, u.Version).Scan(&u.ID)
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

func (m UserModel) DeleteUserById(userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNoSuchUser
	}

	return nil
}

func (m UserModel) UpdateUser(u *User) error {
	stmt := `UPDATE users SET (name, email, password, role, version) =
	($1, $2, $3, $4, version+1)
	WHERE id = $5
	RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, u.Name, u.Email, u.Password.Hashed, u.Role, u.ID).Scan(&u.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil

}
