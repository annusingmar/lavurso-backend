package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type ClassModel struct {
	DB *pgxpool.Pool
}

// DATABASE

func (m ClassModel) InsertClass(c *Class) (*int, error) {
	stmt := `INSERT INTO classes
	(name, teacher_id)
	VALUES
	($1, $2)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	err := m.DB.QueryRow(ctx, stmt, c.Name, c.Teacher.ID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func (m ClassModel) AllClasses() ([]*Class, error) {
	query := `SELECT c.id, c.name, cy.display_name, c.teacher_id, u.name, u.role
	FROM classes c
	INNER JOIN users u
	ON c.teacher_id = u.id
    LEFT JOIN classes_years cy
    ON cy.class_id = c.id AND cy.year_id = (SELECT id FROM years WHERE current is TRUE)`

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
		class.Teacher = new(User)

		err = rows.Scan(
			&class.ID,
			&class.Name,
			&class.DisplayName,
			&class.Teacher.ID,
			&class.Teacher.Name,
			&class.Teacher.Role,
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
	query := `SELECT c.id, c.name, c.teacher_id, u.name, u.role
	FROM classes c
	INNER JOIN users u
	ON c.teacher_id = u.id
	WHERE c.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var class Class
	class.Teacher = new(User)

	err := m.DB.QueryRow(ctx, query, classID).Scan(
		&class.ID,
		&class.Name,
		&class.Teacher.ID,
		&class.Teacher.Name,
		&class.Teacher.Role,
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

func (m ClassModel) GetAllCurrentYearClasses() ([]*Class, error) {
	query := `SELECT c.id, c.name, cy.display_name, c.teacher_id, u.name, u.role
	FROM classes c
	INNER JOIN users u
	ON c.teacher_id = u.id
    INNER JOIN classes_years cy
    ON c.id = cy.class_id AND cy.year_id = (SELECT id FROM years WHERE current is TRUE)`

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
		class.Teacher = new(User)

		err = rows.Scan(
			&class.ID,
			&class.Name,
			&class.DisplayName,
			&class.Teacher.ID,
			&class.Teacher.Name,
			&class.Teacher.Role,
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

func (m ClassModel) GetCurrentYearClassesForTeacher(teacherID int) ([]*Class, error) {
	query := `SELECT c.id, c.name, cy.display_name, c.teacher_id, u.name, u.role
	FROM classes c
	INNER JOIN users u
	ON c.teacher_id = u.id
    INNER JOIN classes_years cy
    ON c.id = cy.class_id AND cy.year_id = (SELECT id FROM years WHERE current is TRUE)
	WHERE c.teacher_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var classes []*Class

	rows, err := m.DB.Query(ctx, query, teacherID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var class Class
		class.Teacher = new(User)

		err = rows.Scan(
			&class.ID,
			&class.Name,
			&class.DisplayName,
			&class.Teacher.ID,
			&class.Teacher.Name,
			&class.Teacher.Role,
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

func (m ClassModel) GetUsersForClassID(classID int) ([]*User, error) {
	query := `SELECT u.id, u.name, u.role
	FROM users u
	WHERE u.class_id = $1
	ORDER BY name ASC`

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
