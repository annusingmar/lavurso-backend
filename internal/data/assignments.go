package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/annusingmar/lavurso-backend/internal/helpers"
	"github.com/annusingmar/lavurso-backend/internal/types"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
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
	ID          int        `json:"id"`
	Journal     *Journal   `json:"journal,omitempty"`
	Subject     *Subject   `json:"subject,omitempty"`
	Description string     `json:"description"`
	Deadline    types.Date `json:"deadline"`
	Type        string     `json:"type"`
	Done        *bool      `json:"done,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type NAssignment struct {
	model.Assignments
	Done    *bool           `json:"done,omitempty" alias:"assignment.done"`
	Subject *model.Subjects `json:"subject,omitempty"`
}

type AssignmentModel struct {
	DB *sql.DB
}

func (m AssignmentModel) GetAssignmentByID(assignmentID int) (*NAssignment, error) {
	query := postgres.SELECT(table.Assignments.AllColumns).
		FROM(table.Assignments).
		WHERE(table.Assignments.ID.EQ(helpers.PostgresInt(assignmentID)))

	var assignment NAssignment

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &assignment)

	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, ErrNoSuchAssignment
		default:
			return nil, err
		}
	}

	return &assignment, nil
}

func (m AssignmentModel) InsertAssignment(a *NAssignment) error {
	stmt := table.Assignments.INSERT(table.Assignments.MutableColumns).
		MODEL(a).
		RETURNING(table.Assignments.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := stmt.QueryContext(ctx, m.DB, a)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) UpdateAssignment(a *NAssignment) error {
	stmt := table.Assignments.UPDATE(table.Assignments.Description, table.Assignments.Deadline, table.Assignments.Type, table.Assignments.UpdatedAt).
		MODEL(a).
		WHERE(table.Assignments.ID.EQ(helpers.PostgresInt(a.ID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) DeleteAssignment(assignmentID int) error {
	stmt := table.Assignments.DELETE().WHERE(table.Assignments.ID.EQ(helpers.PostgresInt(assignmentID)))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) GetAssignmentsByJournalID(journalID int) ([]*model.Assignments, error) {
	query := postgres.SELECT(table.Assignments.AllColumns).
		FROM(table.Assignments).
		WHERE(table.Assignments.JournalID.EQ(helpers.PostgresInt(journalID))).
		ORDER_BY(table.Assignments.Deadline.DESC())

	var assignments []*model.Assignments

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &assignments)
	if err != nil {
		return nil, err
	}

	return assignments, nil
}

func (m AssignmentModel) GetAssignmentsForStudent(studentID int, from, until *types.Date) ([]*NAssignment, error) {
	query := postgres.SELECT(table.Assignments.AllColumns, table.Subjects.AllColumns,
		postgres.CASE().
			WHEN(table.DoneAssignments.UserID.IS_NOT_NULL()).
			THEN(postgres.Bool(true)).
			ELSE(postgres.Bool(false)).
			AS("assignment.done")).
		FROM(table.Assignments.
			INNER_JOIN(table.StudentsJournals, table.StudentsJournals.JournalID.EQ(table.Assignments.JournalID).
				AND(table.StudentsJournals.StudentID.EQ(helpers.PostgresInt(studentID)))).
			INNER_JOIN(table.Journals, table.Journals.ID.EQ(table.Assignments.JournalID)).
			INNER_JOIN(table.Subjects, table.Subjects.ID.EQ(table.Journals.SubjectID)).
			LEFT_JOIN(table.DoneAssignments, table.DoneAssignments.AssignmentID.EQ(table.Assignments.ID).
				AND(table.DoneAssignments.UserID.EQ(table.StudentsJournals.StudentID))))

	if until != nil {
		query = query.WHERE(table.Assignments.Deadline.GT_EQ(postgres.DateT(*from.Time)).
			AND(table.Assignments.Deadline.LT(postgres.DateT(*until.Time))))
	} else {
		query = query.WHERE(table.Assignments.Deadline.GT_EQ(postgres.DateT(*from.Time)))
	}

	query = query.ORDER_BY(table.Assignments.Deadline.ASC())

	var assignments []*NAssignment

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &assignments)
	if err != nil {
		return nil, err
	}

	return assignments, nil
}

func (m AssignmentModel) SetAssignmentDoneForUserID(userID, assignmentID int) error {
	stmt := table.DoneAssignments.INSERT(table.DoneAssignments.AllColumns).
		MODEL(model.DoneAssignments{
			UserID:       &userID,
			AssignmentID: &assignmentID,
		}).
		ON_CONFLICT(table.DoneAssignments.AllColumns...).DO_NOTHING()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m AssignmentModel) RemoveAssignmentDoneForUserID(userID, assignmentID int) error {
	stmt := table.DoneAssignments.DELETE().
		WHERE(table.DoneAssignments.UserID.EQ(helpers.PostgresInt(userID)).
			AND(table.DoneAssignments.AssignmentID.EQ(helpers.PostgresInt(assignmentID))))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
