package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	AssignmentHomework = "homework"
	AssignmentTest     = "test"
)

var (
	ErrNoSuchAssignment     = errors.New("no such assignment")
	ErrNoSuchAssignmentType = errors.New("no such assignment type")
)

type Assignment struct {
	ID          int       `json:"id"`
	JournalID   int       `json:"journal_id"`
	Description string    `json:"description"`
	Deadline    Date      `json:"deadline"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int       `json:"version"`
}

type AssignmentModel struct {
	DB *pgx.Conn
}

func (m AssignmentModel) GetAssignmentByID(assignmentID int) (*Assignment, error) {
	query := `SELECT id, journal_id, description, deadline, type, created_at, updated_at, version
	FROM assignments
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var assignment Assignment

	err := m.DB.QueryRow(ctx, query, assignmentID).Scan(
		&assignment.ID,
		&assignment.JournalID,
		&assignment.Description,
		&assignment.Deadline.Time,
		&assignment.Type,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
		&assignment.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchAssignment
		default:
			return nil, err
		}
	}

	return &assignment, nil
}

func (m AssignmentModel) InsertAssignment(a *Assignment) error {
	stmt := `INSERT INTO assignments
	(journal_id, description, deadline, type, created_at, updated_at, version)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, a.JournalID, a.Description, a.Deadline.Time, a.Type, a.CreatedAt, a.UpdatedAt, a.Version).Scan(&a.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) UpdateAssignment(a *Assignment) error {
	stmt := `UPDATE assignments
	SET (description, deadline, type, updated_at, version)
	= ($1, $2, $3, $4, version+1)
	WHERE id = $5 and version = $6
	RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, a.Description, a.Deadline.Time, a.Type, a.UpdatedAt, a.ID, a.Version).Scan(&a.Version)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m AssignmentModel) DeleteAssignment(assignmentID int) error {
	stmt := `DELETE FROM assignments
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, assignmentID)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) GetAssignmentsByJournalID(journalID int) ([]*Assignment, error) {
	query := `SELECT id, journal_id, description, deadline, type, created_at, updated_at, version
	FROM assignments
	WHERE journal_id = $1
	ORDER BY deadline DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, journalID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var assignments []*Assignment

	for rows.Next() {
		var assignment Assignment
		err = rows.Scan(
			&assignment.ID,
			&assignment.JournalID,
			&assignment.Description,
			&assignment.Deadline.Time,
			&assignment.Type,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
			&assignment.Version,
		)
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, &assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return assignments, nil
}
