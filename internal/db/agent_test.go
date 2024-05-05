package db

import (
	"testing"
	"time"
)

func TestAgentAdd(t *testing.T) {
	sdn := "postgres://postgres:xxiyy@localhost:5432/config"
	if err := Connect(sdn, time.Second); err != nil {
		t.Error(err)
	}
	defer Close()
	if err := AddAgent("default", "test", ServiceStateInit); err != nil {
		t.Error(err)
	}
}

func TestAgentUpdateState(t *testing.T) {
	sdn := "postgres://postgres:xxiyy@localhost:5432/config"
	if err := Connect(sdn, time.Second); err != nil {
		t.Error(err)
	}
	defer Close()
	if err := UpdateAgentState("default", "test", ServiceStateOnline); err != nil {
		t.Error(err)
	}
}
