package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
)

type Log = model.Log

type LogModel struct {
	DB *sql.DB
}

func (m LogModel) InsertLog(l *Log) error {
	stmt := table.Log.INSERT(table.Log.AllColumns).
		MODEL(l)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}
