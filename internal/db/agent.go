package db

import (
	"context"
	"net/url"
	"time"
)

const (
	ServiceStateInit    = "init"
	ServiceStateError   = "error"
	ServiceStateOnline  = "online"
	ServiceStateOffline = "offline"
)

type Agent struct {
	Key       string
	Namespace *Namespace
	Name      string
	State     string
	Metadata  url.Values
	Data      url.Values
	CreatedAt time.Time
	UpdatedAt time.Time
}

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
