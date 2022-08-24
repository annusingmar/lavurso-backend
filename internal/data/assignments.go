package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Journal     *Journal  `json:"journal,omitempty"`
	Subject     *Subject  `json:"subject,omitempty"`
	Description string    `json:"description"`
	Deadline    Date      `json:"deadline"`
	Type        string    `json:"type"`
	Done        *bool     `json:"done,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AssignmentModel struct {
	DB *pgxpool.Pool
}

func (m AssignmentModel) GetAssignmentByID(assignmentID int) (*Assignment, error) {
	query := `SELECT a.id, a.journal_id, j.name, a.description, a.deadline, a.type, a.created_at, a.updated_at
	FROM assignments a
	INNER JOIN journals j
	ON a.journal_id = j.id
	WHERE a.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var assignment Assignment
	assignment.Journal = new(Journal)

	err := m.DB.QueryRow(ctx, query, assignmentID).Scan(
		&assignment.ID,
		&assignment.Journal.ID,
		&assignment.Journal.Name,
		&assignment.Description,
		&assignment.Deadline.Time,
		&assignment.Type,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
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
	(journal_id, description, deadline, type, created_at, updated_at)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, a.Journal.ID, a.Description, a.Deadline.Time, a.Type, a.CreatedAt, a.UpdatedAt).Scan(&a.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) UpdateAssignment(a *Assignment) error {
	stmt := `UPDATE assignments
	SET (description, deadline, type, updated_at)
	= ($1, $2, $3, $4)
	WHERE id = $5`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, a.Description, a.Deadline.Time, a.Type, a.UpdatedAt, a.ID)
	if err != nil {
		return err
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
	query := `SELECT a.id, a.journal_id, j.name, a.description, a.deadline, a.type, a.created_at, a.updated_at
	FROM assignments a
	INNER JOIN journals j
	ON a.journal_id = j.id
	WHERE a.journal_id = $1
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
		assignment.Journal = new(Journal)

		err = rows.Scan(
			&assignment.ID,
			&assignment.Journal.ID,
			&assignment.Journal.Name,
			&assignment.Description,
			&assignment.Deadline.Time,
			&assignment.Type,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
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

func (m AssignmentModel) GetAssignmentsForStudent(studentID int, from, until *Date) ([]*Assignment, error) {
	sqlQuery := `SELECT a.id, s.id, s.name, a.description, a.deadline, a.type,
	(CASE WHEN da.user_id is NOT NULL THEN TRUE ELSE FALSE END),
	a.created_at, a.updated_at
	FROM assignments a
	INNER JOIN users_journals uj
	ON uj.journal_id = a.journal_id
	INNER JOIN journals j
	ON a.journal_id = j.id
	INNER JOIN subjects s
	ON j.subject_id = s.id
	LEFT JOIN done_assignments da
	ON a.id = da.assignment_id AND uj.user_id = da.user_id
	WHERE uj.user_id = $1 AND %s
	ORDER BY a.deadline ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rows pgx.Rows
	var err error

	// can't use parameters for this,
	// so interpolating it directly into query string;
	// but still not trusting user input, so using $2 etc
	if until != nil {
		query := fmt.Sprintf(sqlQuery, "a.deadline >= $2::date AND a.deadline < $3::date")
		rows, err = m.DB.Query(ctx, query, studentID, from.Time, until.Time)
	} else {
		query := fmt.Sprintf(sqlQuery, "a.deadline >= $2::date")
		rows, err = m.DB.Query(ctx, query, studentID, from.Time)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var assignments []*Assignment

	for rows.Next() {
		var assignment Assignment
		assignment.Subject = new(Subject)

		err = rows.Scan(
			&assignment.ID,
			&assignment.Subject.ID,
			&assignment.Subject.Name,
			&assignment.Description,
			&assignment.Deadline.Time,
			&assignment.Type,
			&assignment.Done,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
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

func (m AssignmentModel) SetAssignmentDoneForUserID(userID, assignmentID int) error {
	stmt := `INSERT INTO done_assignments
	(user_id, assignment_id)
	VALUES
	($1, $2)
	ON CONFLICT DO NOTHING`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, assignmentID)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) RemoveAssignmentDoneForUserID(userID, assignmentID int) error {
	stmt := `DELETE FROM done_assignments
	WHERE user_id = $1 and assignment_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, userID, assignmentID)
	if err != nil {
		return err
	}

	return nil
}
