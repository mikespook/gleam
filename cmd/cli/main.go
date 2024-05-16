package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mikespook/gleam/internal/agent"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	ExitCodeDialError    = -3
	ExitCodeConnectError = -5
)

var (
	addr           string
	heartbeat      = time.Second * 10
	ErrWrongFormat = errors.New("wrong format")
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

	c := agent.New("default", "cli", conn)
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
	fmt.Printf("$ exec:(namespace):(key)=(cmd),(args),(async)\n")
	fmt.Printf("$ getconfig:(namespace):(key)\n")
	fmt.Printf("$ getmetadata:(namespace):(key)\n")
	for {
		fmt.Printf("$ ")
		scanner.Scan()
		input := scanner.Text()
		if input == "" {
			continue
		}
		cmd, key, output, err := command(input)
		if err != nil {
			slog.Error("Input", "error", err)
			continue
		}
		switch cmd {
		case "getconfig":
			getconfig(c, key)
		case "getmetadata":
			getmetadata(c, key)
		case "exec":
			exec(c, key, output)
		default:
			slog.Error("Input", "error", "unknown command")
		}
	}
}

func getconfig(c *agent.Client, key *pb.Key) {
	result, err := c.GetConfig(key.Namespace, key.Key)
	if err != nil {
		slog.Error("Config", "error", err)
		return
	}
	slog.Info("Config", "result", result)
}

func getmetadata(c *agent.Client, key *pb.Key) {
	result, err := c.GetMetadata(key.Namespace, key.Key)
	if err != nil {
		slog.Error("Metadata", "error", err)
		return
	}
	slog.Info("Metadata", "result", result)
}

func exec(c *agent.Client, key *pb.Key, input string) {
	reg := regexp.MustCompile(`(\w+)\,?(\w*)\,?(\w*)`)
	matches := reg.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		slog.Debug("Command", "matches", matches, "input", input)
		slog.Error("Exec", "error", "wrong format, require 'command[,args][,async]'")
		return
	}
	var (
		cmd   string
		args  []byte
		async bool
	)
	cmd = matches[0][1]
	args = []byte(matches[0][2])
	async = matches[0][3] == "true"

	result, err := c.Exec(key.Namespace, key.Key, cmd, args, async)
	if err != nil {
		slog.Error("Exec", "error", err)
		return
	}
	slog.Info("Exec", "result", result)
}

func command(input string) (cmd string, key *pb.Key, output string, err error) {
	reg := regexp.MustCompile(`([\w|\-]+):([\w|\-]+):([\w|\-]+)\=?(.*$)`)
	matches := reg.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		err = ErrWrongFormat
		return
	}
	cmd = matches[0][1]
	key = &pb.Key{
		Namespace: matches[0][2],
		Key:       matches[0][3],
	}
	output = matches[0][4]
	slog.Debug("Command", "cmd", cmd, "key", key, "output", output)
	return
}
