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
