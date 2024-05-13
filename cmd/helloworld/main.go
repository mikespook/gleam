package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/mikespook/gleam/internal/agent"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	ExitUnexpectedError  = -1
	ExitCodeDialError    = -3
	ExitCodeConnectError = -5
)

var (
	namespace string
	key       string
	addr      string
	heartbeat = time.Second * 5
	delay     = time.Second * 3
)

func init() {
	flag.StringVar(&namespace, "namespace", "default", "agent namespace")
	flag.StringVar(&key, "key", "helloworld", "agent key")
	flag.StringVar(&addr, "addr", "127.0.0.1:9091", "listen address")
	flag.Parse()

	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(addr, opts)
	if err != nil {
		slog.Error("gRPC", "dial", err)
		os.Exit(ExitCodeDialError)
		return
	}
	defer conn.Close()

	hello := &pb.Command{
		Command: "hello",
	}
	agent.RegisterCommand(hello, helloworld)

	slog.Info("Config", "namespace", namespace, "key", key, "heartbeat", heartbeat, "delay", delay)
	client := agent.New(namespace, key, conn)
	for {
		err := func() (err error) {
			defer func() {
				slog.Warn("Loop", "error", fmt.Sprintf("lost connection, retry in %d seconds", delay/time.Second))
				time.Sleep(delay)
			}()
			if err = client.Loop(heartbeat); err != nil {
				if status.Code(err) == codes.Unavailable {
					err = nil
					return
				}
			}
			return
		}()
		if err != nil {
			slog.Error("Connect", "error", err)
			os.Exit(ExitCodeConnectError)
		}
	}
}

func helloworld(args []byte) ([]byte, error) {
	slog.Info("Command", "hello", args)
	return args, nil
}
