package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoSuchGrade = errors.New("no such grade")
)

type Grade struct {
	ID         *int    `json:"id,omitempty"`
	Identifier *string `json:"identifier,omitempty"`
	Value      *int    `json:"value,omitempty"`
}

type GradeModel struct {
	DB *sql.DB
}

func (m GradeModel) AllGrades() ([]*Grade, error) {
	query := `SELECT id, identifier, value
	FROM grades
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var grades []*Grade

	for rows.Next() {
		var grade Grade
		err := rows.Scan(
			&grade.ID,
			&grade.Identifier,
			&grade.Value,
		)
		if err != nil {
			return nil, err
		}

		grades = append(grades, &grade)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return grades, nil

}

func (m GradeModel) GetGradeByID(gradeID int) (*Grade, error) {
	query := `SELECT id, identifier, value
	FROM grades
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var grade Grade

	err := m.DB.QueryRowContext(ctx, query, gradeID).Scan(
		&grade.ID,
		&grade.Identifier,
		&grade.Value,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoSuchGrade
		default:
			return nil, err
		}
	}

	return &grade, nil
}

func (m GradeModel) UpdateGrade(g *Grade) error {
	stmt := `UPDATE grades
	SET (identifier, value)
	= ($1, $2)
	WHERE id = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, g.Identifier, g.Value, g.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m GradeModel) InsertGrade(g *Grade) error {
	stmt := `INSERT INTO grades
	(identifier, value)
	VALUES
	($1, $2)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, stmt, g.Identifier, g.Value).Scan(&g.ID)
	if err != nil {
		return err
	}

	return nil
}
