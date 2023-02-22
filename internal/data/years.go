package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

var ErrNoCurrentYear = errors.New("no current year set")

type Year = model.Years

type YearExt struct {
	Year
	ClassName *string    `json:"class_name,omitempty" alias:"classes_years.display_name"`
	Stats     *YearStats `json:"stats,omitempty" alias:"stats"`
}

type YearStats struct {
	StudentCount *int `json:"student_count"`
	JournalCount *int `json:"journal_count"`
}

type ClassYear = model.ClassesYears

type YearModel struct {
	DB *sql.DB
}

func (m YearModel) ListAllYears() ([]*YearExt, error) {
	query := postgres.SELECT(table.Years.AllColumns).
		FROM(table.Years)

	var years []*YearExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) ListAllYearsWithStats() ([]*YearExt, error) {
	journalCount := postgres.SELECT(postgres.COUNT(helpers.PostgresInt(1))).
		FROM(table.Journals).
		WHERE(table.Journals.YearID.EQ(table.Years.ID))

	studentCount := postgres.SELECT(postgres.COUNT(helpers.PostgresInt(1))).
		FROM(table.Users.
			INNER_JOIN(table.ClassesYears, table.ClassesYears.YearID.EQ(table.Years.ID))).
		WHERE(table.Users.ClassID.EQ(table.ClassesYears.ClassID))

	query := postgres.SELECT(table.Years.AllColumns, journalCount.AS("stats.journal_count"), studentCount.AS("stats.student_count")).
		FROM(table.Years)

	var years []*YearExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) GetAllYearIDs() ([]int, error) {
	query := postgres.SELECT(table.Years.ID).
		FROM(table.Years)

	var ids []int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (m YearModel) InsertYear(y *Year) error {
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

func (m YearModel) InsertYearForClass(cy *model.ClassesYears) error {
	stmt := table.ClassesYears.INSERT(table.ClassesYears.AllColumns).
		MODEL(cy).
		ON_CONFLICT(table.ClassesYears.ClassID, table.ClassesYears.YearID).
		DO_UPDATE(postgres.SET(table.ClassesYears.DisplayName.SET(postgres.String(*cy.DisplayName))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) RemoveYearsForClass(id int, yearIDs []int) error {
	var yids []postgres.Expression
	for _, id := range yearIDs {
		yids = append(yids, helpers.PostgresInt(id))
	}

	stmt := table.ClassesYears.DELETE().WHERE(table.ClassesYears.ClassID.EQ(helpers.PostgresInt(id)).
		AND(table.ClassesYears.YearID.IN(yids...)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m YearModel) GetCurrentYear() (*YearExt, error) {
	query := postgres.SELECT(table.Years.AllColumns).
		FROM(table.Years).
		WHERE(table.Years.Current.IS_TRUE())

	var year YearExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &year)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, nil
		default:
			return nil, err
		}
	}

	return &year, nil
}

func (m YearModel) GetYearsForStudent(studentID int) ([]*YearExt, error) {

	query := postgres.SELECT(table.Years.ID, table.Years.DisplayName, table.Years.Current, table.ClassesYears.DisplayName).DISTINCT().
		FROM(table.Years.
			INNER_JOIN(table.ClassesYears, table.ClassesYears.YearID.EQ(table.Years.ID)).
			INNER_JOIN(table.Classes, table.Classes.ID.EQ(table.ClassesYears.ClassID)).
			INNER_JOIN(table.Users, table.Users.ClassID.EQ(table.ClassesYears.ClassID))).
		WHERE(table.Users.ID.EQ(helpers.PostgresInt(studentID)))

	var years []*YearExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &years)
	if err != nil {
		return nil, err
	}

	return years, nil
}

func (m YearModel) GetYearsForClass(classID int) ([]*YearExt, error) {
	query := postgres.SELECT(
		table.Years.ID,
		table.Years.DisplayName,
		table.ClassesYears.DisplayName,
	).
		FROM(table.Years.
			LEFT_JOIN(table.ClassesYears, table.ClassesYears.YearID.EQ(table.Years.ID).
				AND(table.ClassesYears.ClassID.EQ(helpers.PostgresInt(classID))))).
		ORDER_BY(table.Years.ID.DESC())

	var cy []*YearExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &cy)
	if err != nil {
		return nil, err
	}

	return cy, nil
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
		WHERE(table.Years.ID.EQ(helpers.PostgresInt(yearID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
