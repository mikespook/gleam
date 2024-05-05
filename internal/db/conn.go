package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	timeout time.Duration
	pool    *pgxpool.Pool
)

func Connect(dsn string, t time.Duration) (err error) {
	timeout = t

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	return Ping()
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return pool.Ping(ctx)
}

func Close() {
	pool.Close()
}
