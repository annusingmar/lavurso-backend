package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	Administrator = iota + 1
	Parent
	Student
)

var (
	ErrEmailAlreadyExists = errors.New("an user with specified email already exists")
)

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
