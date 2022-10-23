package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Year struct {
	ID          int        `json:"id"`
	DisplayName string     `json:"display_name"`
	Courses     *int       `json:"courses"`
	Current     bool       `json:"current"`
	Stats       *YearStats `json:"stats"`
}

type YearStats struct {
	StudentCount *int `json:"student_count"`
	JournalCount *int `json:"journal_count"`
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
	query := `SELECT id, display_name, courses, current
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
			&year.Current,
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

func (m YearModel) ListAllYearsWithStats() ([]*Year, error) {
	query := `SELECT y.id, y.display_name, y.courses, y.current,
	(
	  SELECT COUNT(1)
	  FROM journals j
	  WHERE j.year_id = y.id
	),
	(
	  SELECT COUNT(1)
	  FROM users u
	  INNER JOIN classes_years cy
	  ON cy.year_id = y.id
	  WHERE u.class_id = cy.class_id
	)
	FROM years y`

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
		year.Stats = new(YearStats)

		err = rows.Scan(
			&year.ID,
			&year.DisplayName,
			&year.Courses,
			&year.Current,
			&year.Stats.JournalCount,
			&year.Stats.StudentCount,
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

func (m YearModel) InsertYear(y *Year) (*int, error) {
	stmt := `INSERT INTO years
	(display_name, courses)
	VALUES ($1, $2)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	err := m.DB.QueryRow(ctx, stmt, y.DisplayName, y.Courses).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (m YearModel) InsertYearForClass(cy *ClassYear) error {
	stmt := `INSERT INTO classes_years
	(class_id, year_id, display_name)
	VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, cy.ClassID, cy.YearID, cy.DisplayName)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) GetCurrentYear() (*Year, error) {
	query := `SELECT id, display_name, courses, current
	FROM years
	WHERE current is TRUE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var year Year

	err := m.DB.QueryRow(ctx, query).Scan(
		&year.ID,
		&year.DisplayName,
		&year.Courses,
		&year.Current,
	)
	if err != nil {
		return nil, err
	}

	return &year, err
}

func (m YearModel) GetYearsForStudent(studentID int) ([]*Year, error) {
	query := `SELECT DISTINCT y.id, y.display_name, y.courses, y.current
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
			&year.Current,
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

func (m YearModel) RemoveCurrentYear() error {
	stmt := `UPDATE years
	SET current = false`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) SetYearAsCurrent(yearID int) error {
	stmt := `UPDATE years
	SET current = true
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, yearID)
	if err != nil {
		return err
	}

	return nil
}
