package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mikespook/gleam/internal/agent"
)

const (
	ExitCodeDialError    = -3
	ExitCodeConnectError = -5
)

var (
	addr      string
	heartbeat = time.Second * 10
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:9091", "")
	flag.Parse()

	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(addr, opts)
	if err != nil {
		slog.Error("gRPC", "error", err)
		os.Exit(ExitCodeDialError)
		return
	}
	defer conn.Close()

	c, err := agent.New("default", "cli", conn)
	if err != nil {
		slog.Error("Client", "error", err)
		os.Exit(ExitCodeConnectError)
	}

	defer func() {
		if err := c.Close(); err != nil {
			slog.Error("Client", "error", err)
			os.Exit(ExitCodeConnectError)
		}
	}()

	agent.Wait(conn)

	go func() {
		if err := c.Loop(heartbeat); err != nil {
			slog.Error("Connect", "error", err)
			os.Exit(ExitCodeConnectError)
		}
	}()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("========")
		fmt.Printf("Namespace: ")
		scanner.Scan()
		namespace := scanner.Text()
		fmt.Printf("Key: ")
		scanner.Scan()
		key := scanner.Text()
		fmt.Printf("Command: ")
		scanner.Scan()
		command := scanner.Text()
		fmt.Printf("Args: ")
		scanner.Scan()
		args := scanner.Text()
		fmt.Printf("Async [Y/N]: ")
		scanner.Scan()
		asyncStr := scanner.Text()
		async := asyncStr == "Y"
		fmt.Printf("Namespace=%s Key=%s Command=%s Args=%s Async=%v\n", namespace, key, command, args, async)
		result, err := c.Exec(namespace, key, command, []byte(args), async)
		if err != nil {
			slog.Error("Command", "error", err)
			continue
		}
		slog.Info("Command", "result", result)
	}

}
