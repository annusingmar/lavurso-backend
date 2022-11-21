package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

type Year struct {
	ID          int        `json:"id"`
	DisplayName string     `json:"display_name"`
	Courses     *int       `json:"courses"`
	Current     bool       `json:"current"`
	Stats       *YearStats `json:"stats"`
}

type NYear struct {
	model.Years
	Stats *YearStats `json:"stats,omitempty" alias:"stats.*"`
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
	DB *sql.DB
}

func (m YearModel) ListAllYears() ([]*NYear, error) {
	query := postgres.SELECT(table.Years.AllColumns).
		FROM(table.Years)

	fmt.Println(query.DebugSql())

	var years []*NYear

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) ListAllYearsWithStats() ([]*NYear, error) {
	journalCount := postgres.SELECT(postgres.COUNT(postgres.Int32(int32(1)))).
		FROM(table.Journals).
		WHERE(table.Journals.YearID.EQ(table.Years.ID))

	studentCount := postgres.SELECT(postgres.COUNT(postgres.Int32(int32(1)))).
		FROM(table.Users.
			INNER_JOIN(table.ClassesYears, table.ClassesYears.YearID.EQ(table.Years.ID))).
		WHERE(table.Users.ClassID.EQ(table.ClassesYears.ClassID))

	query := postgres.SELECT(table.Years.AllColumns, journalCount.AS("stats.journal_count"), studentCount.AS("stats.student_count")).
		FROM(table.Years)

	var years []*NYear

	fmt.Println(query.DebugSql())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) InsertYear(y *Year) (*int, error) {
	stmt := `INSERT INTO years
	(display_name, courses, current)
	VALUES ($1, $2, false)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	err := m.DB.QueryRowContext(ctx, stmt, y.DisplayName, y.Courses).Scan(&id)
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

	_, err := m.DB.ExecContext(ctx, stmt, cy.ClassID, cy.YearID, cy.DisplayName)
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

	err := m.DB.QueryRowContext(ctx, query).Scan(
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

	rows, err := m.DB.QueryContext(ctx, query, studentID)
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

	_, err := m.DB.ExecContext(ctx, stmt)
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

	_, err := m.DB.ExecContext(ctx, stmt, yearID)
	if err != nil {
		return err
	}

	return nil
}
