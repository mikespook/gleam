package db

import (
	"testing"
	"time"

	"github.com/mikespook/gleam/internal/pb"
)

func TestAddAgent(t *testing.T) {
	sdn := "postgres://postgres:123456@localhost:5432/config"
	if err := Connect(sdn, time.Second); err != nil {
		t.Error(err)
	}
	defer Close()
	state := ServiceStateInit
	key := &pb.Key{
		Namespace: "default",
		Key:       "test",
		State:     &state,
	}
	daemon := &pb.Key{
		Namespace: "daemon",
		Key:       "test",
	}
	if err := AddAgent(key, daemon); err != nil {
		t.Error(err)
	}
}

func TestUpdateAgentState(t *testing.T) {
	sdn := "postgres://postgres:123456@localhost:5432/config"
	if err := Connect(sdn, time.Second); err != nil {
		t.Error(err)
	}
	defer Close()
	state := ServiceStateOnline
	key := &pb.Key{
		Namespace: "default",
		Key:       "test",
		State:     &state,
	}
	daemon := &pb.Key{
		Namespace: "daemon",
		Key:       "test",
	}
	if err := UpdateAgentState(key, daemon); err != nil {
		t.Error(err)
	}
}

func TestGetAgentMetadata(t *testing.T) {
	sdn := "postgres://postgres:123456@localhost:5432/config"
	if err := Connect(sdn, time.Second); err != nil {
		t.Error(err)
	}
	defer Close()
	key := &pb.Key{
		Namespace: "default",
		Key:       "test",
	}
	metadata, err := GetAgentMetadata(key)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Metadata: %s", metadata)
}
