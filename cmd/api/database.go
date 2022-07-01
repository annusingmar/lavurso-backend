package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func openDBConnection(connectionString string) *pgxpool.Pool {
	pool, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		log.Fatalln(err)
	}
	return pool
}
