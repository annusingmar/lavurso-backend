package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoClassForUser = errors.New("no class set for user")
	ErrNoSuchClass    = errors.New("no such class")
)

type Class struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	TeacherID int    `json:"teacher_id"`
	Archived  bool   `json:"archived"`
}

type ClassModel struct {
	DB *sql.DB
}

func (m ClassModel) GetClassByID(classID int) (*Class, error) {
	query := `SELECT id, name, teacher_id, archived
	FROM classes
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var class Class

	err := m.DB.QueryRowContext(ctx, query, classID).Scan(
		&class.ID,
		&class.Name,
		&class.TeacherID,
		&class.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchClass
		default:
			return nil, err
		}
	}

	return &class, nil

}

func (m ClassModel) GetClassForUserID(userID int) (*Class, error) {
	query := `SELECT id, name, teacher_id, archived
	FROM classes c
	INNER JOIN users_classes uc
	ON uc.class_id = c.id
	WHERE uc.user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var class Class

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&class.ID,
		&class.Name,
		&class.TeacherID,
		&class.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoClassForUser
		default:
			return nil, err
		}
	}

	return &class, nil

}

func (m ClassModel) GetUsersForClassID(classID int) ([]*User, error) {
	query := `SELECT id, name, email, password, role, created_at, active, version
	FROM users u
	INNER JOIN users_classes uc
	ON uc.user_id = u.id
	WHERE uc.class_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, classID)
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
