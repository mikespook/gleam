package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
	heartbeat = time.Second * 10
)

func init() {
	flag.StringVar(&namespace, "namespace", "default", "")
	flag.StringVar(&key, "key", "helloworld", "")
	flag.StringVar(&addr, "addr", "127.0.0.1:9091", "")
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

	client, err := agent.New(namespace, key, conn)
	if err != nil {
		slog.Error("Connect", "error", err)
		os.Exit(ExitCodeConnectError)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			slog.Error("Connect", "error", err)
			os.Exit(ExitUnexpectedError)
		}
	}()
	hello := &pb.Command{
		Command: "hello",
	}
	agent.RegisterCommand(hello, helloworld)
	if err := client.Loop(heartbeat); err != nil {
		slog.Error("Connect", "error", err)
		os.Exit(ExitCodeConnectError)
	}
}

func helloworld(args []byte) ([]byte, error) {
	slog.Info("Command", "hello", args)
	return args, nil
}
