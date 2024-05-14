package db

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/pb"
)

func AddDaemon(key *pb.Key) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.daemons(namespace, key, name)
	VALUES ($1, $2, $3)
	ON CONFLICT (namespace, key) DO UPDATE SET updated_at = now();`
	tag, err := pool.Exec(ctx, sql, key.Namespace, key.Key, key.Key)
	if err != nil {
		return err
	}
	slog.Debug("DB", "daemon-add", tag)
	return nil
}

func HeartbeatDaemon(key *pb.Key) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.daemons SET updated_at=now() WHERE namespace=$1 and key=$2;`
	tag, err := pool.Exec(ctx, sql, key.Namespace, key.Key)
	if err != nil {
		return err
	}
	slog.Debug("DB", "daemon-heartbeat", tag)
	return nil
}

func SetDaemonAddr(key *pb.Key, addr string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.daemons SET updated_at = now(), metadata = jsonb_set(metadata, '{addr}', $3) WHERE namespace = $1 and key = $2;`
	var addrBytes []byte
	addrBytes, err = json.Marshal(addr)
	tag, err := pool.Exec(ctx, sql, key.Namespace, key.Key, addrBytes)
	if err != nil {
		return err
	}
	slog.Debug("DB", "daemon-set-addr", tag)
	return nil
}

func GetDaemonMetadata(key *pb.Key) (metadata []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT metadata FROM public.daemons WHERE namespace = $1 and key = $2;"
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

func GetDaemonAddr(key *pb.Key) (addr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT metadata->>'addr' FROM public.daemons WHERE namespace = $1 and key = $2;"
	rows, err := pool.Query(ctx, sql, key.Namespace, key.Key)
	if err != nil {
		return
	}
	addrs, err := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (addr string, err error) {
		if err = row.Scan(&addr); err != nil {
			return
		}
		return
	})
	if len(addrs) == 0 {
		err = pgx.ErrNoRows
		return
	}
	return
}

func CleanDaemon(key *pb.Key) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.agents SET updated_at=now(), state='offline',
	daemon_namespace=null, daemon_key=null WHERE daemon_namespace=$1 and daemon_key=$2;`
	tag, err := pool.Exec(ctx, sql, key.Namespace, key.Key)
	if err != nil {
		return err
	}
	slog.Debug("DB", "daemon-clean", tag)
	return nil
}
