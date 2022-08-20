package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotValidAbsence = errors.New("not valid absence for user")
	ErrNoSuchAbsence   = errors.New("no such absence")
	ErrAbsenceExcused  = errors.New("absence already excused")
	ErrNoSuchExcuse    = errors.New("no such excuse")
)

type AbsenceExcuse struct {
	ID     *int       `json:"id"`
	MarkID *int       `json:"mark_id"`
	Excuse *string    `json:"excuse"`
	By     *User      `json:"by"`
	At     *time.Time `json:"at"`
}

type AbsenceModel struct {
	DB *pgxpool.Pool
}

func (m AbsenceModel) InsertExcuse(excuse *AbsenceExcuse) error {
	stmt := `INSERT INTO absences_excuses
	(absence_mark_id, excuse, by, at)
	VALUES
	($1, $2, $3, $4)
	RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, excuse.MarkID, excuse.Excuse, excuse.By, excuse.At).Scan(&excuse.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m AbsenceModel) DeleteExcuse(excuseID int) error {
	stmt := `DELETE FROM absences_excuses
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, stmt, excuseID)
	if err != nil {
		return err
	}

	return nil
}

func (m AbsenceModel) GetExcuseByID(excuseID int) (*AbsenceExcuse, error) {
	query := `SELECT id, absence_mark_id, excuse, by, at
	FROM absences_excuses
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var excuse AbsenceExcuse

	err := m.DB.QueryRow(ctx, query, excuseID).Scan(
		&excuse.ID,
		&excuse.MarkID,
		&excuse.Excuse,
		&excuse.By,
		&excuse.At,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNoSuchExcuse
		default:
			return nil, err
		}
	}

	return &excuse, nil
}
