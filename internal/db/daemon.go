package db

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	row := pool.QueryRow(ctx, sql, key.Namespace, key.Key)
	err = row.Scan(&addr)
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

type Daemon struct {
	pb.Key
	Name      string    `json:"name"`
	Meatadata string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetDaemons() (agents []*Daemon, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `SELECT namespace, key, name, metadata, created_at, updated_at FROM public.daemons;`
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*Daemon, error) {
		var namespace, key, name pgtype.Text
		d := &Daemon{}
		err := row.Scan(&namespace, &key, &name, &d.Meatadata, &d.CreatedAt, &d.UpdatedAt)
		d.Namespace = namespace.String
		d.Key.Key = key.String
		d.Name = name.String
		return d, err
	})
}
