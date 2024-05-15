package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	ServiceStateInit    = "init"
	ServiceStateError   = "error"
	ServiceStateOnline  = "online"
	ServiceStateOffline = "offline"
)

func AddAgent(key, daemon *pb.Key) error {
	var daemonNamespace, daemonKey *string
	if daemon != nil {
		daemonNamespace = &daemon.Namespace
		daemonKey = &daemon.Key
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.agents(namespace, key, name, state, daemon_namespace, daemon_key)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (namespace, key) DO UPDATE SET updated_at = now(), daemon_namespace = $5, daemon_key = $6;`
	tag, err := pool.Exec(ctx, sql, key.Namespace, key.Key, key.Key, key.State, daemonNamespace, daemonKey)
	if err != nil {
		return err
	}
	slog.Debug("DB", "agent-add", tag)
	return nil
}

func UpdateAgentState(key, daemon *pb.Key) error {
	var daemonNamespace, daemonKey *string
	if daemon != nil {
		daemonNamespace = &daemon.Namespace
		daemonKey = &daemon.Key
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.agents SET updated_at=now(),
	state=$1, daemon_namespace=$2, daemon_key=$3 WHERE namespace=$4 and key=$5;`
	tag, err := pool.Exec(ctx, sql, key.State, daemonNamespace, daemonKey, key.Namespace, key.Key)
	if err != nil {
		return err
	}
	slog.Debug("DB", "agent-update", tag)
	return nil
}

func GetAgentMetadata(key *pb.Key) (metadata []byte, err error) {
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

type Agent struct {
	State string
	pb.Key
}

func GetAgentState(key *pb.Key) (agent *Agent, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := "SELECT state, daemon_namespace, daemon_key FROM public.agents WHERE namespace = $1 and key = $2;"
	row := pool.QueryRow(ctx, sql, key.Namespace, key.Key)
	var state string
	var agentNamespace, agentKey pgtype.Text
	if err = row.Scan(&state, &agentNamespace, &agentKey); err != nil {
		return
	}
	agent = &Agent{
		State: state,
		Key: pb.Key{
			Namespace: agentNamespace.String,
			Key:       agentKey.String,
		},
	}
	return
}
