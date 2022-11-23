package data

import (
	"context"
	"database/sql"
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

type ClassYear = model.ClassesYears

type YearModel struct {
	DB *sql.DB
}

func (m YearModel) ListAllYears() ([]*NYear, error) {
	query := postgres.SELECT(table.Years.AllColumns).
		FROM(table.Years)

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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) InsertYear(y *model.Years) error {
	stmt := table.Years.INSERT(table.Years.MutableColumns).
		MODEL(y).RETURNING(table.Years.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, y)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) InsertYearForClass(cy *ClassYear) error {
	stmt := table.ClassesYears.INSERT(table.ClassesYears.AllColumns).
		MODEL(cy)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) GetCurrentYear() (*NYear, error) {
	query := postgres.SELECT(table.Years.AllColumns).
		FROM(table.Years).
		WHERE(table.Years.Current.IS_TRUE())

	var year NYear

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &year)

	return &year, err
}

func (m YearModel) GetYearsForStudent(studentID int) ([]*NYear, error) {

	query := postgres.SELECT(table.Years.ID, table.Years.DisplayName, table.Years.Courses, table.Years.Current).DISTINCT().
		FROM(table.Years.
			INNER_JOIN(table.ClassesYears, table.ClassesYears.YearID.EQ(table.Years.ID)).
			INNER_JOIN(table.Classes, table.Classes.ID.EQ(table.ClassesYears.ClassID)).
			INNER_JOIN(table.Users, table.Users.ClassID.EQ(table.ClassesYears.ClassID))).
		WHERE(table.Users.ID.EQ(postgres.Int32(int32(studentID))))

	var years []*NYear

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) RemoveCurrentYear() error {
	stmt := table.Years.UPDATE(table.Years.Current).
		SET(postgres.Bool(false)).
		WHERE(table.Years.Current.IS_TRUE())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) SetYearAsCurrent(yearID int) error {
	stmt := table.Years.UPDATE(table.Years.Current).
		SET(postgres.Bool(true)).
		WHERE(table.Years.ID.EQ(postgres.Int32(int32(yearID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
