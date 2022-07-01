package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
)

type AbsenceExcuse struct {
	ID            *int       `json:"id"`
	AbsenceMarkID *int       `json:"absence_id"`
	Excuse        *string    `json:"excuse"`
	By            *int       `json:"by"`
	At            *time.Time `json:"at"`
}

type AbsenceModel struct {
	DB *pgx.Conn
}

func (m AbsenceModel) GetAbsenceMarksByUserID(userID int) ([]*Mark, error) {
	query := `SELECT
	m.id, m.user_id, m.lesson_id, m.course, m.journal_id, m.grade_id, m.subject_id, m.comment, m.type, m.by, m.at, exc.id, exc.absence_mark_id, exc.excuse, exc.by, exc.at
	FROM marks m
	LEFT JOIN absences_excuses exc
	ON m.id = exc.absence_mark_id
	WHERE m.user_id = $1 and m.type = 'absent'`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var marks []*Mark

	for rows.Next() {
		var mark Mark
		mark.AbsenceExcuses = new(AbsenceExcuse)

		err := rows.Scan(
			&mark.ID,
			&mark.UserID,
			&mark.LessonID,
			&mark.Course,
			&mark.JournalID,
			&mark.GradeID,
			&mark.SubjectID,
			&mark.Comment,
			&mark.Type,
			&mark.By,
			&mark.At,
			&mark.AbsenceExcuses.ID,
			&mark.AbsenceExcuses.AbsenceMarkID,
			&mark.AbsenceExcuses.Excuse,
			&mark.AbsenceExcuses.By,
			&mark.AbsenceExcuses.At,
		)
		if err != nil {
			return nil, err
		}

		if mark.AbsenceExcuses.AbsenceMarkID == nil {
			mark.AbsenceExcuses = nil
		}

		marks = append(marks, &mark)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return marks, nil
}
