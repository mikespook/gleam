package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	// Receive the initial agent key from the client
	var agentKey *pb.Key
	if agentKey, err = s.recv(nil, stream); err != nil {
		slog.Error("Connect", "recv", err)
		return
	}
	//
	var a *db.Agent
	if a, err = db.GetAgentState(agentKey); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("Connect", "error", err)
		return
	}
	if a != nil && a.State != db.ServiceStateOffline {
		err = ErrDuplicatedAgent
		slog.Warn("Connect", "error", err)
		return
	}
	var agent *Agent
	// Add the agent to the server's agents collection
	if agent, err = s.addAgent(agentKey, s.key); err != nil {
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
		if err = s.setAgent(key, s.key); err != nil {
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
