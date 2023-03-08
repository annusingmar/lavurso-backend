package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/model"
	"github.com/annusingmar/lavurso-backend/internal/data/gen/lavurso/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

type Log = model.Logs

type LogExt struct {
	Log
	User  *User `json:"user,omitempty"`
	Total *int  `json:"-"`
}

type LogsWithTotal struct {
	Logs  []*LogExt `json:"logs,omitempty"`
	Total int       `json:"total"`
}

type LogModel struct {
	DB *sql.DB
}

func (m LogModel) InsertLog(l *Log) error {
	stmt := table.Logs.INSERT(table.Logs.MutableColumns).
		MODEL(l)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := stmt.ExecContext(ctx, m.DB)
	if err != nil {
		return err
	}

	return nil
}

func (m LogModel) AllLogs(page, limit int, search string) (*LogsWithTotal, error) {
	query := table.Logs.SELECT(postgres.COUNT(postgres.STAR).OVER().AS("logext.total"), table.Logs.AllColumns, table.Users.ID, table.Users.Name).
		FROM(table.Logs.
			LEFT_JOIN(table.Users, table.Users.ID.EQ(table.Logs.UserID))).
		ORDER_BY(table.Logs.At.DESC()).
		GROUP_BY(table.Logs.ID, table.Users.ID)

	if search != "" {
		s := postgres.LOWER(postgres.String("%" + search + "%"))
		query = query.WHERE(postgres.OR(
			postgres.LOWER(table.Users.Name).LIKE(s),
			postgres.LOWER(table.Logs.Target).LIKE(s),
		))
	}

	if limit != 0 {
		offset := (page - 1) * limit
		query = query.OFFSET(int64(offset)).LIMIT(int64(limit))
	}

	var logs []*LogExt

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := query.QueryContext(ctx, m.DB, &logs)
	if err != nil {
		return nil, err
	}

	l := &LogsWithTotal{
		Logs: logs,
	}

	if len(l.Logs) > 0 {
		l.Total = *l.Logs[0].Total
	} else {
		l.Total = 0
	}

	return l, nil
}
