package databaseconn

import (
	"chacalc/src"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connection() *pgxpool.Pool {
	// TODO: replace this and put it in the the .env
	db_url := "postgres://deeznutz:0000@localhost:5433/chacalc"
	conn, err := pgxpool.New(context.Background(), db_url)
	if err != nil {
		src.Logger.Fatalf("there was an issue reaching the database %v", err.Error())
	}
	return conn
}
