package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/pb"
)

func GetConfig(key *pb.Key) (value []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT value FROM public.configs WHERE namespace = $1 and key = $2;"
	rows, err := pool.Query(ctx, sql, key.Namespace, key.Key)
	if err != nil {
		return nil, err
	}
	value, err = pgx.CollectOneRow(rows, func(row pgx.CollectableRow) ([]byte, error) {
		var metadata []byte
		err := row.Scan(metadata)
		return metadata, err
	})
	return
}
