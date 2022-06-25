package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func openDBConnection(connectionString string) *sql.DB {
	pool, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatalln(err)
	}
	if err = pool.Ping(); err != nil {
		log.Fatalln(err)
	}
	return pool
}
