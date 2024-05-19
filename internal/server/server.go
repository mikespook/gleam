package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/pb"
)

var (
	// ErrDuplicatedAgent is an error indicating that an attempt was made to add
	// an agent that already exists, resulting in a duplication.
	ErrDuplicatedAgent = errors.New("duplicated agent")
	// ErrMissingAgent is an error indicating that an agent was not found or is
	// missing when performing an operation that requires its presence.
	ErrMissingAgent = errors.New("missing agent")
	// ErrMissingDaemon is an error indicating that an agent was not found or is
	// missing when performing an operation that requires its presence.
	ErrMissingDaemon = errors.New("missing daemon")
)

// Serve creates a new gRPC server and starts serving on the provided listener.
// It registers the Gleam server implementation with the gRPC server.
// The function logs a message when the server is started.
//
// Parameters:
//   - addr: The network address to listen on (e.g., "localhost:50051").
//   - l: A net.Listener representing the listening socket.
//
// Returns an error if there was an issue starting the gRPC server.
func Serve(namespace, key string, l net.Listener, heartbeat time.Duration) (err error) {
	if key == "" {
		if key, err = os.Hostname(); err != nil {
			return
		}
	}
	daemonKey := &pb.Key{
		Namespace: namespace,
		Key:       key,
	}
	if err = db.AddDaemon(daemonKey); err != nil {
		return
	}
	if err = db.SetDaemonAddr(daemonKey, l.Addr().String()); err != nil {
		return
	}
	if err = db.CleanDaemon(daemonKey); err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go deamonHeartbeats(ctx, daemonKey, time.Second)
	s := grpc.NewServer()
	pb.RegisterGleamServer(s, newServer(daemonKey, heartbeat))
	return s.Serve(l)
}

func deamonHeartbeats(ctx context.Context, key *pb.Key, t time.Duration) {
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			db.HeartbeatDaemon(key)
		}
	}
}

// Server represents the gRPC server for the Gleam service.
// It embeds pb.UnimplementedGleamServer to implement the GleamServer interface.
// The server includes a sync.RWMutex for managing concurrent access to the agents.
// The agents field holds the collection of agents managed by the server.
type Server struct {
	pb.UnimplementedGleamServer

	agents      Agents
	agentMutex  sync.RWMutex
	daemons     Daemons
	daemonMutex sync.RWMutex
	heartbeat   time.Duration

	key *pb.Key
}

func newServer(key *pb.Key, heartbeat time.Duration) *Server {
	s := &Server{
		agents:    newAgents(),
		daemons:   newDaemons(),
		heartbeat: heartbeat,
		key:       key,
	}
	return s
}

func (s *Server) GetMetadata(ctx context.Context, key *pb.Key) (value *pb.Value, err error) {
	var metadataValue []byte
	if metadataValue, err = db.GetAgentMetadata(key); err != nil {
		return
	}
	value = &pb.Value{
		Value: metadataValue,
	}
	return
}

func (s *Server) GetConfig(ctx context.Context, key *pb.Key) (value *pb.Value, err error) {
	var configValue []byte
	if configValue, err = db.GetConfig(key); err != nil {
		return
	}
	value = &pb.Value{
		Value: configValue,
	}
	return
}

// Exec executes a command for the specified agent.
//
// Parameters:
//   - ctx: The context for the execution.
//   - cmd: The command to execute.
//
// Returns:
//   - value: The result of the command execution.
//   - err: An error if any operation encounters an error.
func (s *Server) Exec(ctx context.Context, cmd *pb.Command) (value *pb.Value, err error) {
	// Send the command to the agent's command channel
	var a *Agent
	// check agent locally
	a, err = s.getAgent(cmd.Key)
	if err != nil && !errors.Is(err, ErrMissingAgent) {
		return
	}
	if errors.Is(err, ErrMissingAgent) {
		var agentDB *db.Agent
		agentDB, err = db.GetAgentState(cmd.Key)
		if err != nil {
			return
		}
		if agentDB == nil || agentDB.State != db.ServiceStateOnline {
			err = ErrMissingAgent
			return
		}

		var daemon *Daemon
		if daemon, err = s.getDaemon(agentDB); errors.Is(err, ErrMissingDaemon) {
			var addr string
			addr, err = db.GetDaemonAddr(&agentDB.Daemon)
			if err != nil {
				return
			}
			if addr == "" {
				err = ErrMissingDaemon
				return
			}

			opts := grpc.WithTransportCredentials(insecure.NewCredentials())
			var conn *grpc.ClientConn
			conn, err = grpc.Dial(addr, opts)
			if err != nil {
				return nil, err
			}
			daemon = newDaemon(agentDB.Daemon.Namespace, agentDB.Daemon.Key, conn)
			go func() {
				defer daemon.Close()
				if err := daemon.Loop(s.heartbeat); err != nil {
					slog.Error("Exec", "error", err)
				}
			}()
			s.setDaemon(daemon)
		}
		if err != nil {
			return
		}
		var result []byte
		result, err = daemon.Client.Exec(cmd.Key.Namespace, cmd.Key.Key, cmd.Command, cmd.Args, *cmd.Async)
		if err != nil {
			value = nil
			return
		}
		value = &pb.Value{
			Value: result,
		}
		return
	}

	if err = db.AddCommand(cmd); err != nil {
		return
	}
	a.cmds <- cmd
	return
}

// Callback handles the callback from the worker for an executed command.
// It updates the database with the result and returns the corresponding value.
func (s *Server) Callback(ctx context.Context, result *pb.Result) (value *pb.Value, err error) {
	slog.Debug("Callback", "result", result, "uuid", result.Uuid)
	if err = db.UpdateCommand(result); err != nil {
		slog.Error("Callback", "error", err)
		return
	}
	return
}
