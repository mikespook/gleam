package db

import (
	"context"
	"log/slog"

	"github.com/mikespook/gleam/internal/pb"
)

const (
	CommandStateSuccess  = "success"
	CommandStateError    = "error"
	CommandStateNotFound = "not_found"
	CommandStateTimeout  = "timeout"
	CommandStateRunning  = "running"
)

type Command struct {
	ID      string
	Sender  *Agent
	Worker  *Agent
	Command string
	Args    []byte
	Async   bool
}

func AddCommand(cmd *Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `INSERT INTO public.stream_logs(uuid, command, args, async)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (uuid) DO UPDATE SET updated_at = now();`
	_, err := pool.Exec(ctx, sql, cmd.ID, cmd.Command, cmd.Args, cmd.Async)
	return err
}

func UpdateCommand(result *pb.Result) error {
	state := CommandStateSuccess
	if result.IsError {
		state = CommandStateError
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	sql := `UPDATE public.stream_logs
	SET result=$1, state=$2, updated_at=now()
	WHERE uuid = $3;`
	_, err := pool.Exec(ctx, sql, result.Value, state, result.Uuid)
	slog.Debug("update", "uuid", result.Uuid, "error", err)
	return err
}
