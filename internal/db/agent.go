package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const (
	ServiceStateInit    = "init"
	ServiceStateError   = "error"
	ServiceStateOnline  = "online"
	ServiceStateOffline = "offline"
)

func AddAgent(namespace, key, state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.agents(namespace, key, name, state, metadata)
	VALUES ($1, $2, $3, $4, '{}')
	ON CONFLICT (namespace, key) DO UPDATE SET updated_at = now();`
	_, err := pool.Exec(ctx, sql, namespace, key, key, state)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAgentState(namespace, key, state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.agents SET updated_at=now(),
	state=$1 WHERE namespace=$2 and key=$3;`
	_, err := pool.Exec(ctx, sql, state, namespace, key)
	if err != nil {
		return err
	}
	return nil
}

func GetMetadata(key string) (metadata []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT metadata FROM public.agents WHERE key = $1;"
	rows, err := pool.Query(ctx, sql, key)
	if err != nil {
		return nil, err
	}
	metadata, err = pgx.CollectOneRow(rows, func(row pgx.CollectableRow) ([]byte, error) {
		var metadata []byte
		err := row.Scan(metadata)
		return metadata, err
	})
	return
}
