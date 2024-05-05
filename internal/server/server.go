package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
func Serve(l net.Listener) (err error) {
	s := grpc.NewServer()
	pb.RegisterGleamServer(s, newServer())
	slog.Info("Server", "serve", l.Addr())
	return s.Serve(l)
}

// Server represents the gRPC server for the Gleam service.
// It embeds pb.UnimplementedGleamServer to implement the GleamServer interface.
// The server includes a sync.RWMutex for managing concurrent access to the agents.
// The agents field holds the collection of agents managed by the server.
type Server struct {
	sync.RWMutex
	pb.UnimplementedGleamServer

	agents Agents
}

func newServer() *Server {
	s := &Server{
		agents: newAgents(),
	}
	return s
}

func (s *Server) addAgent(key *pb.Key) (agent *Agent, err error) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.agents[key.Namespace]; !ok {
		s.agents[key.Namespace] = newAgentsGrouped()
	}
	if _, ok := s.agents[key.Namespace][key.Key]; ok {
		err = ErrDuplicatedAgent
		return
	}
	if err = db.AddAgent(key.Namespace, key.Key, db.ServiceStateInit); err != nil {
		return
	}
	agent = newAgent(key.Namespace, key.Key)
	s.agents[key.Namespace][key.Key] = agent
	return
}

func (s *Server) setAgent(key *pb.Key) (err error) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.agents[key.Namespace]; !ok {
		return
	}
	if _, ok := s.agents[key.Namespace][key.Key]; !ok {
		return
	}
	s.agents[key.Namespace][key.Key].UpdatedAt = time.Now()
	s.agents[key.Namespace][key.Key].State = *key.State
	return db.UpdateAgentState(key.Key, key.Namespace, *key.State)
}

func (s *Server) removeAgent(key *pb.Key) (err error) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.agents[key.Namespace]; !ok {
		return
	}
	if _, ok := s.agents[key.Namespace][key.Key]; !ok {
		return
	}
	delete(s.agents[key.Namespace], key.Key)
	return db.UpdateAgentState(key.Key, key.Namespace, db.ServiceStateOffline)
}

// Connect handles the connection with a client and manages the agent's lifecycle.
// It receives the initial agent key from the client, adds the agent to the server's agents collection,
// and starts the command and heartbeat loops for the agent.
//
// Parameters:
//   - stream: The stream for receiving and sending messages with the client.
//
// Returns:
//   - err: An error if any operation encounters an error.
func (s *Server) Connect(stream pb.Gleam_ConnectServer) (err error) {
	// Recover from panics to handle errors gracefully
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic: %v", r)
			}
			slog.Error("Connect", "error", err)
			slog.Debug("Connect", "stack", string(debug.Stack()))
		}
	}()
	var agentKey *pb.Key
	// Receive the initial agent key from the client
	if agentKey, err = s.recv(nil, stream); err != nil {
		slog.Error("Connect", "recv", err)
		return
	}
	var agent *Agent
	// Add the agent to the server's agents collection
	if agent, err = s.addAgent(agentKey); err != nil {
		slog.Error("Connect", "namespace", agentKey.Namespace, "key", agentKey.Key, "error", err)
		return
	}
	// Defer removing the agent when Connect returns
	defer func() {
		if err = s.removeAgent(agentKey); err != nil {
			slog.Error("Connect", "remove_agent", err)
		}
	}()

	slog.Debug("Connect", "namespace", agentKey.Namespace,
		"key", agentKey.Key, "state", *agentKey.State)

	// Create a context for the agent's lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	// Start the command loop for the agent in a separate goroutine
	go s.commandLoop(ctx, agent, stream)

	// Start the heartbeat loop for the agent
	s.heartbeatLoop(cancel, agent, stream)
	return
}

func (s *Server) commandLoop(ctx context.Context, agent *Agent, stream pb.Gleam_ConnectServer) {
	for {
		select {
		case <-ctx.Done():
			slog.Debug("Command", "canceled", ctx.Err().Error())
			return
		case cmd := <-agent.cmds:
			if err := stream.Send(cmd); err != nil {
				slog.Error("Command", "error", err)
				return
			}
			slog.Debug("Command", "detail", cmd)
		}
	}
}

func (s *Server) heartbeatLoop(cancel context.CancelFunc, agent *Agent, stream pb.Gleam_ConnectServer) {
	defer cancel()
	for {
		key, err := s.recv(agent, stream)
		if err != nil {
			return
		}
		if err = s.setAgent(key); err != nil {
			slog.Error("Heartbeat", "set_agent", err)
			return
		}
		slog.Debug("Heartbeat", "namespace", agent.Namespace, "key",
			agent.Key, "state", *key.State)
	}
}

func (s *Server) recv(agent *Agent, stream pb.Gleam_ConnectServer) (key *pb.Key, err error) {
	if agent == nil {
		agent = newAgent("unknown", "unknown")
	}
	key, err = stream.Recv()
	if err == io.EOF {
		slog.Info("Disconnect", "namespace", agent.Namespace, "key", agent.Key)
		return
	}
	switch status.Code(err) {
	case codes.OK:
	case codes.Canceled:
		slog.Warn("Abnormal Disconnect", "namespace", agent.Namespace, "key", agent.Key)
		return
	default:
		slog.Error("Abnormal Disconnect with Error",
			"namespace", agent.Namespace, "key", agent.Key, "error", err)
		return
	}
	return
}

func (s *Server) GetMetadata(ctx context.Context, key *pb.Key) (value *pb.Value, err error) {
	return
}

func (s *Server) GetConfig(ctx context.Context, key *pb.Key) (value *pb.Value, err error) {
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
	// Check if the agent exists in the server's agents collection
	if _, ok := s.agents[cmd.Key.Namespace]; !ok {
		err = ErrMissingAgent
		return
	}
	if _, ok := s.agents[cmd.Key.Namespace][cmd.Key.Key]; !ok {
		err = ErrMissingAgent
		return
	}
	obj := &db.Command{
		ID:      cmd.Uuid,
		Command: cmd.Command,
		Args:    cmd.Args,
	}
	if err = db.AddCommand(obj); err != nil {
		return
	}
	// Send the command to the agent's command channel
	s.agents[cmd.Key.Namespace][cmd.Key.Key].cmds <- cmd
	return
}

func (s *Server) Callback(ctx context.Context, result *pb.Result) (value *pb.Value, err error) {
	slog.Debug("Callback", "result", result, "uuid", result.Uuid)
	if err = db.UpdateCommand(result); err != nil {
		slog.Error("Callback", "error", err)
		return
	}
	return
}
