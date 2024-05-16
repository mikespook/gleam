package server

import (
	"github.com/mikespook/gleam/internal/agent"
	"github.com/mikespook/gleam/internal/db"
	"google.golang.org/grpc"
)

type Daemon struct {
	*agent.Client
}

func newDaemon(namespace, key string, conn grpc.ClientConnInterface) *Daemon {
	return &Daemon{
		Client: agent.New(namespace, key, conn),
	}
}

type DaemonsGrouped map[string]*Daemon

func newDaemonsGrouped() DaemonsGrouped {
	return make(DaemonsGrouped)
}

type Daemons map[string]DaemonsGrouped

func newDaemons() Daemons {
	return make(Daemons)
}

func (s *Server) setDaemon(daemon *Daemon) {
	s.daemonMutex.Lock()
	defer s.daemonMutex.Unlock()
	if _, ok := s.daemons[daemon.Namespace]; !ok {
		s.daemons[daemon.Namespace] = newDaemonsGrouped()
	}
	s.daemons[daemon.Namespace][daemon.Key.Key] = daemon
}

func (s *Server) getDaemon(agent *db.Agent) (daemon *Daemon, err error) {
	s.daemonMutex.RLock()
	defer s.daemonMutex.RUnlock()
	// Check if the agent exists in the server's agents collection
	if _, ok := s.daemons[agent.Daemon.Namespace]; !ok {
		return nil, ErrMissingDaemon
	}
	if _, ok := s.daemons[agent.Daemon.Namespace][agent.Daemon.Key]; !ok {
		return nil, ErrMissingDaemon
	}
	return s.daemons[agent.Daemon.Namespace][agent.Daemon.Key], nil
}
