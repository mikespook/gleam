package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	ServiceStateInit    = "init"
	ServiceStateError   = "error"
	ServiceStateOnline  = "online"
	ServiceStateOffline = "offline"
)

func AddAgent(key, daemon *pb.Key) error {
	if daemon == nil {
		daemon = &pb.Key{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.agents(namespace, key, name, state, daemon_namespace, deamon_key)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (namespace, key) DO UPDATE SET updated_at = now(), daemon_namespace = $5, daemon_key = $6;`
	tag, err := pool.Exec(ctx, sql, key.Namespace, key.Key, key.Key, key.State, daemon.Namespace, daemon.Key)
	if err != nil {
		return err
	}
	slog.Debug("DB", "agent-add", tag)
	return nil
}

func UpdateAgentState(key, daemon *pb.Key) error {
	if daemon == nil {
		daemon = &pb.Key{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.agents SET updated_at=now(),
	state=$1, daemon_namespace=$2, deamon_key=$3 WHERE namespace=$4 and key=$5;`
	tag, err := pool.Exec(ctx, sql, key.State, daemon.Namespace, daemon.Key, key.Namespace, key.Key)
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
	sql := "SELECT state, daemon_namespace, deamon_key FROM public.agents WHERE namespace = $1 and key = $2;"
	rows, err := pool.Query(ctx, sql, key.Namespace, key.Key)
	if err != nil {
		return nil, err
	}
	agents, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*Agent, error) {
		agent := &Agent{}
		if err := row.Scan(&agent.State, &agent.Namespace, &agent.Key); err != nil {
			return nil, err
		}
		return agent, nil
	})
	if len(agents) == 0 {
		return nil, nil
	}
	return agents[0], nil
}
