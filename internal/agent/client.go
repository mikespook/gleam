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

type Client struct {
	Namespace string
	Key       string
	cc        pb.GleamClient
	stream    pb.Gleam_ConnectClient
}

func New(namespace, key string, conn grpc.ClientConnInterface) (client *Client) {
	client = &Client{
		Namespace: namespace,
		Key:       key,
	}
	client.cc = pb.NewGleamClient(conn)
	return
}

func (client *Client) Loop(heartbeat time.Duration) (err error) {
	client.stream, err = register(client.Namespace, client.Key, client.cc)
	if err != nil {
		return
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go client.loop(cancel)
	return client.heartbeats(ctx, heartbeat)
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

func (client *Client) GetConfig(namespace, key string) (result []byte, err error) {
	pbkey := &pb.Key{
		Namespace: namespace,
		Key:       key,
	}
	v, err := client.cc.GetConfig(context.Background(), pbkey)
	if err != nil {
		return nil, err
	}
	return v.Value, nil
}
func (client *Client) GetMetadata(namespace, key string) (result []byte, err error) {
	pbkey := &pb.Key{
		Namespace: namespace,
		Key:       key,
	}
	v, err := client.cc.GetMetadata(context.Background(), pbkey)
	if err != nil {
		return nil, err
	}
	return v.Value, nil
}

func (client *Client) loop(cancel context.CancelFunc) (err error) {
	defer cancel()
	var cmd *pb.Command
	for {
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if _, err = client.cc.Callback(ctx, result); err != nil {
			slog.Error("Exec", "uuid", cmd.Uuid, "error", err)
			return
		}
	}
}

func (client *Client) heartbeats(ctx context.Context, t time.Duration) (err error) {
	defer client.stream.CloseSend()
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	if err = client.beat(); err != nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err = client.beat(); err != nil {
				return
			}
		}
	}
	return
}

func (client *Client) beat() (err error) {
	state := db.ServiceStateOnline
	key := pb.Key{Namespace: client.Namespace, Key: client.Key, State: &state}
	return client.stream.Send(&key)
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
