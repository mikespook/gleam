package db

import (
	"testing"
	"time"

	"github.com/mikespook/gleam/internal/pb"
)

func TestConfig(t *testing.T) {
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
	if err := AddAgent(key); err != nil {
		t.Error(err)
	}
}
