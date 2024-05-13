package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/server"
)

const (
	ExitCodeServeError = -1
)

var (
	addr           string
	dbDSN          string
	logLevel       string
	namespace, key string
	dbTimeout      = time.Second * 5
)

func init() {
	flag.StringVar(&logLevel, "log-level", "info", "log level in [debug, info, warn, error]")
	flag.StringVar(&dbDSN, "dsn", "postgres://postgres@localhost:5432/config", "database DSN")
	flag.StringVar(&addr, "addr", "127.0.0.1:9091", "listen address")
	flag.StringVar(&namespace, "namespace", "gleamd", "daemon namespace")
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Flag", "error", err)
	}
	flag.StringVar(&key, "key", hostname, "daemon key")
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
	if err := dbInit(); err != nil {
		slog.Error("DB", "init", err)
		return
	}
	slog.Info("DB", "database", "initialised")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Network", "listen", err)
		return
	}
	slog.Info("Server", "listen", addr)

	if err := server.Serve(namespace, key, l); err != nil {
		slog.Error("Server", "serve", err)
		os.Exit(ExitCodeServeError)
	}
	slog.Info("Exit")
}

func dbInit() (err error) {
	return db.AddNamespace(db.DefaultNamespace, db.DefaultNamespaceName)
}
