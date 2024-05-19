package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/agent"
	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/pb"
)

const (
	ExitCodeDialError    = -3
	ExitCodeConnectError = -5

	DefaultDSN      = "postgres://postgres@localhost:5432/config"
	DefaultAddr     = "127.0.0.1:9091"
	DefaultHTTPAddr = "127.0.0.1:8080"
	DefaultLogLevel = "info"
	DefaultKey      = "webd"
	heartbeat       = time.Second * 10
	dbTimeout       = time.Second * 5
)

var (
	addr, httpAddr, dbDSN, logLevel string
	ErrWrongFormat                  = errors.New("wrong format")
)

func init() {
	flag.StringVar(&logLevel, "log-level", DefaultLogLevel, "log level in [debug, info, warn, error]")
	flag.StringVar(&dbDSN, "dsn", DefaultDSN, "database DSN")
	flag.StringVar(&addr, "addr", DefaultAddr, "The network address for the main server.")
	flag.StringVar(&httpAddr, "http-addr", DefaultHTTPAddr, "The network address for the HTTP server.")
	flag.Parse()

	switch logLevel {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	default: // info
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	envDSN := os.Getenv("POSTGRESQL_DSN")
	if envDSN != "" {
		dbDSN = envDSN
	}
}

func main() {
	if err := db.Connect(dbDSN, dbTimeout); err != nil {
		slog.Error("DB", "connect", err)
		return
	}

	config, _ := pgx.ParseConfig(dbDSN)
	slog.Info("DB", "connected", fmt.Sprintf("%s@%s:%d/%s", config.User, config.Host, config.Port, config.Database))

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(addr, opts)
	if err != nil {
		slog.Error("gRPC", "error", err)
		os.Exit(ExitCodeDialError)
		return
	}
	defer conn.Close()

	c := agent.New(db.DefaultNamespace, DefaultKey, conn)
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
	indexHandler, err := newIndexHandler()
	if err != nil {
		slog.Error("HTTP", "error", err)
		os.Exit(ExitCodeDialError)
		return
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/exec", execHandler)
	http.HandleFunc("/config", configHandler)
	http.HandleFunc("/metadata", metadataHandler)
	http.HandleFunc("/agents", agentsHandler)
	http.HandleFunc("/daemons", daemonsHandler)
	if err = http.ListenAndServe(httpAddr, nil); err != nil {
		slog.Error("Connect", "error", err)
		os.Exit(ExitCodeConnectError)
	}
	slog.Info("Exit")
}
func newIndexHandler() (http.HandlerFunc, error) {
	f, err := os.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	page, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write(page)
	}
	return h, nil
}

func execHandler(w http.ResponseWriter, r *http.Request) {

}

func configHandler(w http.ResponseWriter, r *http.Request) {

}

func metadataHandler(w http.ResponseWriter, r *http.Request) {

}

func agentsHandler(w http.ResponseWriter, r *http.Request) {
	agents, err := db.GetAgents()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
	jsonBytes, err := json.Marshal(agents)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
	if _, err := w.Write(jsonBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
}

func daemonsHandler(w http.ResponseWriter, r *http.Request) {
	daemons, err := db.GetDaemons()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
	jsonBytes, err := json.Marshal(daemons)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
	if _, err := w.Write(jsonBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToBytes(err))
		return
	}
}

func errToBytes(err error) []byte {
	return []byte(fmt.Sprintf("{\"error\":%s}", strconv.Quote(err.Error())))
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
