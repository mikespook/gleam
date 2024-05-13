package server

import (
	"time"

	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/pb"
)

type Agent struct {
	Namespace string
	Key       string
	UpdatedAt time.Time
	State     string

	cmds chan *pb.Command
}

func (a *Agent) Send(cmd *pb.Command) {
	a.cmds <- cmd
}

func newAgent(namespace, key string) *Agent {
	return &Agent{
		Namespace: namespace,
		Key:       key,
		UpdatedAt: time.Now(),
		State:     db.ServiceStateInit,
		cmds:      make(chan *pb.Command, 16),
	}
}

type AgentsGrouped map[string]*Agent

func newAgentsGrouped() AgentsGrouped {
	return make(AgentsGrouped)
}

type Agents map[string]AgentsGrouped

func newAgents() Agents {
	return make(Agents)
}

func (s *Server) addAgent(key, daemon *pb.Key) (agent *Agent, err error) {
	s.agentMutex.Lock()
	defer s.agentMutex.Unlock()
	if _, ok := s.agents[key.Namespace]; !ok {
		s.agents[key.Namespace] = newAgentsGrouped()
	}
	*key.State = db.ServiceStateInit
	if err = db.AddAgent(key, daemon); err != nil {
		return
	}
	agent = newAgent(key.Namespace, key.Key)
	s.agents[key.Namespace][key.Key] = agent
	return
}

func (s *Server) setAgent(key, daemon *pb.Key) (err error) {
	s.agentMutex.Lock()
	defer s.agentMutex.Unlock()
	if _, ok := s.agents[key.Namespace]; !ok {
		err = ErrMissingAgent
		return
	}
	if _, ok := s.agents[key.Namespace][key.Key]; !ok {
		err = ErrMissingAgent
		return
	}
	s.agents[key.Namespace][key.Key].UpdatedAt = time.Now()
	s.agents[key.Namespace][key.Key].State = *key.State
	return db.UpdateAgentState(key, daemon)
}

func (s *Server) removeAgent(key *pb.Key) (err error) {
	s.agentMutex.Lock()
	defer s.agentMutex.Unlock()
	if _, ok := s.agents[key.Namespace]; !ok {
		return
	}
	if _, ok := s.agents[key.Namespace][key.Key]; !ok {
		return
	}
	delete(s.agents[key.Namespace], key.Key)
	*key.State = db.ServiceStateOffline
	return db.UpdateAgentState(key, nil)
}

func (s *Server) getAgent(key *pb.Key) (a *Agent, err error) {
	s.agentMutex.RLock()
	defer s.agentMutex.RUnlock()
	if a, ok := s.agents[key.Namespace][key.Key]; ok {
		return a, nil
	}
	return nil, ErrMissingAgent
}
