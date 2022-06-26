package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoClassForUser = errors.New("no class set for user")
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
