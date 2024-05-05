package agent

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/mikespook/gleam/internal/pb"
)

var (
	mutex    sync.RWMutex
	commands = make(map[string]CommandFunc)

	ErrCommandNotFound = errors.New("command not found")
)

type CommandFunc func([]byte) ([]byte, error)

func RegisterCommand(cmd *pb.Command, f CommandFunc) {
	mutex.Lock()
	defer mutex.Unlock()
	commands[cmd.Command] = f
}

func execCommand(cmd *pb.Command) (result []byte, err error) {
	mutex.RLock()
	defer mutex.RUnlock()
	f, ok := commands[cmd.Command]
	if !ok {
		err = ErrCommandNotFound
		return
	}
	if *cmd.Async {
		// async exec
		go func() {
			if _, err = f(cmd.Args); err != nil {
				slog.Error("Exec", "uuid", cmd.Uuid, "error", err)
				return
			}
		}()
	}
	// sync exec
	if result, err = f(cmd.Args); err != nil {
		slog.Error("Exec", "uuid", cmd.Uuid, "error", err)
		return
	}
	return
}
