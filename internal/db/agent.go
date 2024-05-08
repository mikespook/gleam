package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	ServiceStateInit    = "init"
	ServiceStateError   = "error"
	ServiceStateOnline  = "online"
	ServiceStateOffline = "offline"
)

func AddAgent(key *pb.Key) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.agents(namespace, key, name, state, metadata)
	VALUES ($1, $2, $3, $4, '{}')
	ON CONFLICT (namespace, key) DO UPDATE SET updated_at = now();`
	_, err := pool.Exec(ctx, sql, key.Namespace, key.Key, key.Key, key.State)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAgentState(key *pb.Key) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.agents SET updated_at=now(),
	state=$1 WHERE namespace=$2 and key=$3;`
	_, err := pool.Exec(ctx, sql, key.State, key.Namespace, key.Key)
	if err != nil {
		return err
	}
	return nil
}

func GetMetadata(key *pb.Key) (metadata []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT metadata FROM public.agents WHERE namespace = $1 and key = $2;"
	rows, err := pool.Query(ctx, sql, key.Namespace, key.Key)
	if err != nil {
		return nil, err
	}
	metadata, err = pgx.CollectOneRow(rows, func(row pgx.CollectableRow) ([]byte, error) {
		v := row.RawValues()
		return v[0], err
	})
	return
}
