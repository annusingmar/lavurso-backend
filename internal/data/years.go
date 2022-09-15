package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Year struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Courses     *int   `json:"courses"`
}

type ClassYear struct {
	ClassID     int    `json:"class_id"`
	YearID      int    `json:"year_id"`
	DisplayName string `json:"display_name"`
}

type YearModel struct {
	DB *pgxpool.Pool
}

func (m YearModel) ListAllYears() ([]*Year, error) {
	query := `SELECT id, display_name, courses
	FROM years`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var years []*Year

	for rows.Next() {
		var year Year

		err = rows.Scan(
			&year.ID,
			&year.DisplayName,
			&year.Courses,
		)
		if err != nil {
			return nil, err
		}

		years = append(years, &year)
	}

	if rows.Err(); err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) InsertYear(y *Year) error {
	stmt := `INSERT INTO years
	(display_name, courses)
	VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, y.DisplayName, y.Courses)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) GetCurrentYear() (*Year, error) {
	query := `SELECT id, display_name, courses
	FROM years
	WHERE current is TRUE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var year Year

	err := m.DB.QueryRow(ctx, query).Scan(
		&year.ID,
		&year.DisplayName,
		&year.Courses,
	)
	if err != nil {
		return nil, err
	}

	return &year, err
}

func (m YearModel) GetYearsForStudent(studentID int) ([]*Year, error) {
	query := `SELECT DISTINCT y.id, y.display_name, y.courses
	FROM years y
	INNER JOIN classes_years cy
	ON y.id = cy.year_id
	INNER JOIN classes c
	ON c.id = cy.class_id
	INNER JOIN users u
	ON u.class_id = cy.class_id
	WHERE u.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var years []*Year

	for rows.Next() {
		var year Year

		err = rows.Scan(
			&year.ID,
			&year.DisplayName,
			&year.Courses,
		)
		if err != nil {
			return nil, err
		}

		years = append(years, &year)
	}

	if rows.Err(); err != nil {
		return nil, err
	}

	return years, nil
}
