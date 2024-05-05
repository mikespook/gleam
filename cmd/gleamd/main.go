package main

import (
	"flag"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/mikespook/gleam/internal/db"
	"github.com/mikespook/gleam/internal/server"
)

const (
	ExitCodeServeError = -1
)

var (
	addr      string
	dbDSN     string
	dbTimeout = time.Second * 5
)

func init() {
	flag.StringVar(&dbDSN, "dsn", "postgres://postgres:xxiyy@localhost:5432/config", "database DSN")
	flag.StringVar(&addr, "addr", "127.0.0.1:9091", "listen address")
	flag.Parse()

	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func main() {
	if err := db.Connect(dbDSN, dbTimeout); err != nil {
		slog.Error("DB", "connect", err)
		return
	}
	slog.Info("Main", "database", "connected")
	if err := dbInit(); err != nil {
		slog.Error("DB", "init", err)
		return
	}
	slog.Info("Main", "database", "initialised")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Network", "listen", err)
		return
	}
	slog.Info("Main", "listen", addr)
	if err := server.Serve(l); err != nil {
		slog.Error("Server", "serve", err)
		os.Exit(ExitCodeServeError)
	}
	slog.Info("Exit")
}

func dbInit() (err error) {
	return db.AddNamespace(db.DefaultNamespace, db.DefaultNamespaceName)
}
