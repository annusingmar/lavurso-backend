package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDBConnection(connectionString string) *sql.DB {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		log.Fatalln(err)
	}
	return db
}
