package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mikespook/golib/log"
	"github.com/mikespook/golib/pid"
	"github.com/mikespook/golib/signal"

	"github.com/mikespook/gleam"
)

const (
	SCRIPT_ROOT = "GLEAM_SCRIPT_ROOT"
)

var (
	etcd, id, region, script, pidfile string
)

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
		return
	}

	if !flag.Parsed() {
		flag.StringVar(&etcd, "etcd", "http://127.0.0.1:7001", "url of etcd")
		flag.StringVar(&id, "id", fmt.Sprintf("%s-%d", hostname, os.Getpid()), "node id")
		flag.StringVar(&region, "region", "default", "regions to watch, multi-regions splite by `:`")
		flag.StringVar(&script, "script", "", "directory of lua scripts")
		flag.StringVar(&pidfile, "pid", "", "PID file")

		flag.Parse()
	}
	log.InitWithFlag()

	if script == "" {
		script = os.Getenv(SCRIPT_ROOT)
	}
	if script == "" {
		script = "./"
	}
}

func main() {
	if pidfile != "" {
		p, err := pid.New(pidfile)
		if err != nil {
			log.Error(err)
			return
		}
		defer p.Close()
	}

	log.Message("Starting...")

	g, err := gleam.New(id, script)
	if err != nil {
		log.Error(err)
		return
	}
	defer g.Close()

	log.Messagef("Watching(ID = %s)...", id)
	if err := g.Watch(gleam.MakeNode(id)); err != nil {
		log.Error(err)
		return
	}
	for _, r := range strings.Split(region, ":") {
		log.Messagef("Watching(Region = %s)...", r)
		if err := g.Watch(gleam.MakeRegion(r)); err != nil {
			log.Error(err)
			return
		}
	}
	go g.Serve()

	// signal handler
	sh := signal.NewHandler()
	sh.Bind(os.Interrupt, func() bool { return true })
	sh.Loop()
	log.Message("Exit!")
}
