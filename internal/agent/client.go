package agent

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/google/uuid"
	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	Namespace = "default"
	Key       = "default"
)

var ()

type Client struct {
	Namespace string
	Key       string
	cc        pb.GleamClient
	stream    pb.Gleam_ConnectClient
}

func New(namespace, key string, conn grpc.ClientConnInterface) (client *Client, err error) {
	c := &Client{
		Namespace: namespace,
		Key:       key,
	}
	c.cc = pb.NewGleamClient(conn)
	c.stream, err = register(namespace, key, c.cc)
	if err != nil {
		return
	}
	client = c
	return
}

func (client *Client) Loop(heartbeat time.Duration) error {
	ctx, cancle := context.WithCancel(context.Background())
	go client.heartbeats(cancle, heartbeat)
	return client.loop(ctx)
}

func (client *Client) Close() error {
	return client.stream.CloseSend()
}

func (client *Client) Exec(namespace, key, command string, args []byte, async bool) (result []byte, err error) {
	cmd := &pb.Command{
		Uuid: uuid.New().String(),
		Key: &pb.Key{
			Namespace: namespace,
			Key:       key,
		},
		Command: command,
		Args:    args,
		Async:   &async,
	}
	v, err := client.cc.Exec(context.Background(), cmd)
	if err != nil {
		return nil, err
	}
	return v.Value, nil
}

func Wait(conn *grpc.ClientConn) {
	for {
		state := conn.GetState()
		slog.Info("Connection", "state", state)
		if state == connectivity.Ready {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func register(namespace, agentKey string, cc pb.GleamClient) (stream pb.Gleam_ConnectClient, err error) {
	stream, err = cc.Connect(context.Background())
	if err != nil {
		return
	}

	state := db.ServiceStateInit
	key := pb.Key{Namespace: namespace, Key: agentKey, State: &state}
	err = stream.Send(&key)
	return
}

func (client *Client) loop(ctx context.Context) (err error) {
	var cmd *pb.Command
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if cmd, err = client.stream.Recv(); err != nil {
				return
			}
			result := &pb.Result{
				Uuid: cmd.Uuid,
			}
			if result.Value, err = execCommand(cmd); err != nil {
				result.Value = []byte(err.Error())
				result.IsError = true
			}
			slog.Debug("Exec", "cmd", cmd)
			slog.Debug("Exec", "result", result)
			if _, err = client.cc.Callback(ctx, result); err != nil {
				slog.Error("Exec", "uuid", cmd.Uuid, "error", err)
				return
			}
		}
	}
}

func (client *Client) heartbeats(cancel context.CancelFunc, t time.Duration) (err error) {
	defer cancel()
	defer client.stream.CloseSend()
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	for range ticker.C {
		state := db.ServiceStateOnline
		key := pb.Key{Namespace: Namespace, Key: Key, State: &state}
		if err = client.stream.Send(&key); err != nil {
			return
		}
	}
	return
}
