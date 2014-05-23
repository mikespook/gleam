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
	CONFIG_FILE = "GLEAM_CONFIG"
)

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
		return
	}
	var configFile, etcd, caFile, certFile, keyFile, name, region, script, pidFile string
	if !flag.Parsed() {
		flag.StringVar(&configFile, "config", "", "Path to configuration file")
		flag.StringVar(&etcd, "etcd", "127.0.0.1:7001", "URL of etcd")
		flag.StringVar(&caFile, "ca-file", "", "Path to the CA file")
		flag.StringVar(&certFile, "cert-file", "", "Path to the cert file")
		flag.StringVar(&keyFile, "key-file", "", "Path to the ke:y file")
		flag.StringVar(&name, "name", fmt.Sprintf("%s-%d", hostname, os.Getpid()), "Name of this node")
		flag.StringVar(&region, "region", "default", "Regions to watch, multi-regions splite by `:`")
		flag.StringVar(&script, "script", "", "Directory of lua scripts")
		flag.StringVar(&pidFile, "pid", "", "PID file")

		flag.Parse()
	}
	log.InitWithFlag()

	var config *Config

	if configFile == "" {
		configFile = os.Getenv(CONFIG_FILE)
	}

	if configFile != "" {
		var err error
		if config, err = LoadConfig(configFile); err != nil {
			log.Error(err)
			return
		}
	} else {
		config = &Config{
			Name:   name,
			Pid:    pidFile,
			Script: script,
			Region: strings.Split(region, ":"),
			Etcd: ConfigEtcd{
				Url:  etcd,
				Ca:   caFile,
				Cert: certFile,
				Key:  keyFile,
			},
		}
	}

	if config.Pid != "" {
		p, err := pid.New(config.Pid)
		if err != nil {
			log.Error(err)
			return
		}
		defer p.Close()
	}

	log.Message("Starting...")

	g, err := gleam.New(config.Name, config.Script)
	if err != nil {
		log.Error(err)
		return
	}
	defer g.Close()

	log.Messagef("Watching(Name = %s)...", config.Name)
	if err := g.Watch(gleam.MakeNode(config.Name)); err != nil {
		log.Error(err)
		return
	}
	for _, r := range config.Region {
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
