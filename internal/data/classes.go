package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNoClassForUser = errors.New("no class set for user")
	ErrNoSuchClass    = errors.New("no such class")
	ErrClassArchived  = errors.New("class is archived")
)

type Class struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Teacher  *User  `json:"teacher"`
	Archived bool   `json:"archived"`
}

type ClassModel struct {
	DB *pgxpool.Pool
}

// DATABASE

func (m ClassModel) InsertClass(c *Class) error {
	stmt := `INSERT INTO classes
	(name, teacher_id, archived)
	VALUES
	($1, $2, $3)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, c.Name, c.Teacher.ID, c.Archived).Scan(&c.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m ClassModel) AllClasses() ([]*Class, error) {
	query := `SELECT c.id, c.name, c.teacher_id, u.name, u.role, c.archived
	FROM classes c
	INNER JOIN users u
	ON c.teacher_id = u.id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var classes []*Class

	rows, err := m.DB.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var class Class
		class.Teacher = &User{}

		err = rows.Scan(
			&class.ID,
			&class.Name,
			&class.Teacher.ID,
			&class.Teacher.Name,
			&class.Teacher.Role,
			&class.Archived,
		)

		if err != nil {
			return nil, err
		}

		classes = append(classes, &class)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return classes, nil
}

func (m ClassModel) GetAllClassIDs() ([]int, error) {
	query := `SELECT
	array(SELECT id	FROM classes)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ids []int

	err := m.DB.QueryRow(ctx, query).Scan(&ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m ClassModel) UpdateClass(c *Class) error {
	stmt := `UPDATE classes SET (name, teacher_id) =
	($1, $2)
	WHERE id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, c.Name, c.Teacher.ID, c.ID)
	if err != nil {
		return err
	}

	return nil

}

func (m ClassModel) GetClassByID(classID int) (*Class, error) {
	query := `SELECT c.id, c.name, c.teacher_id, u.name, u.role, c.archived
	FROM classes c
	INNER JOIN users u
	ON c.teacher_id = u.id
	WHERE c.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var class Class
	class.Teacher = &User{}

	err := m.DB.QueryRow(ctx, query, classID).Scan(
		&class.ID,
		&class.Name,
		&class.Teacher.ID,
		&class.Teacher.Name,
		&class.Teacher.Role,
		&class.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchClass
		default:
			return nil, err
		}
	}

	return &class, nil

}

func (m ClassModel) GetClassForUserID(userID int) (*Class, error) {
	query := `SELECT c.id, c.name, c.teacher_id, u.name, u.role, c.archived
	FROM classes c
	INNER JOIN users_classes uc
	ON uc.class_id = c.id
	INNER JOIN users u
	ON c.teacher_id = u.id
	WHERE uc.user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var class Class
	class.Teacher = &User{}

	err := m.DB.QueryRow(ctx, query, userID).Scan(
		&class.ID,
		&class.Name,
		&class.Teacher.ID,
		&class.Teacher.Name,
		&class.Teacher.Role,
		&class.Archived,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoClassForUser
		default:
			return nil, err
		}
	}

	return &class, nil

}

func (m ClassModel) GetUsersForClassID(classID int) ([]*User, error) {
	query := `SELECT id, name, role
	FROM users u
	INNER JOIN users_classes uc
	ON uc.user_id = u.id
	WHERE uc.class_id = $1
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, classID)
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

func (m ClassModel) SetClassIDForUserID(userID, classID int) error {
	stmt := `INSERT INTO users_classes
	(user_id, class_id)
	VALUES ($1, $2)
	ON CONFLICT (user_id)
	DO UPDATE SET class_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, classID)
	if err != nil {
		return err
	}
	return nil
}

func (m ClassModel) IsUserInClass(userID, classID int) (bool, error) {
	query := `SELECT COUNT(1) FROM users_classes
	WHERE user_id = $1 and class_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, userID, classID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (m ClassModel) DoesParentHaveChildInClass(parentID, classID int) (bool, error) {
	query := `SELECT COUNT(1)
	FROM parents_children pc
	INNER JOIN users_classes uc
	ON pc.child_id = uc.user_id
	WHERE pc.parent_id = $1
	AND uc.class_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var result int

	err := m.DB.QueryRow(ctx, query, parentID, classID).Scan(&result)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}
