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
	// prepare the configuration
	config, err := InitConfig()
	if err != nil {
		log.Error(err)
		return
	}

	// make PID file
	if config.Pid != "" {
		p, err := pid.New(config.Pid)
		if err != nil {
			log.Error(err)
			return
		}
		defer p.Close()
	}

	log.Message("Starting...")

	g, err := gleam.New(config.Etcd, config.Name, config.Script)
	if err != nil {
		log.Error(err)
		return
	}
	defer g.Close()
	if config.Ca != "" || config.Cert != "" || config.Key != "" {
		g.TLS(config.Cert, config.Key, config.Ca)
	}

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

func InitConfig() (*Config, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	var configFile, etcd, caFile, certFile, keyFile, name, region, script, pidFile string
	if !flag.Parsed() {
		flag.StringVar(&configFile, "config", "", "Path to configuration file")
		flag.StringVar(&etcd, "etcd", "127.0.0.1:4001", "A comma-delimited list of etcd")
		flag.StringVar(&caFile, "ca-file", "", "Path to the CA file")
		flag.StringVar(&certFile, "cert-file", "", "Path to the cert file")
		flag.StringVar(&keyFile, "key-file", "", "Path to the ke:y file")
		flag.StringVar(&name, "name", fmt.Sprintf("%s-%d", hostname, os.Getpid()), "Name of this node")
		flag.StringVar(&region, "region", "default", "A comma-delimited list of regions to watch")
		flag.StringVar(&script, "script", "", "Directory of lua scripts")
		flag.StringVar(&pidFile, "pid", "", "PID file")

		flag.Parse()
	}
	log.InitWithFlag()

	config := Config{
		Name:   name,
		Pid:    pidFile,
		Script: script,
		Region: strings.Split(region, ","),
		Etcd:   strings.Split(etcd, ","),
		Ca:     caFile,
		Cert:   certFile,
		Key:    keyFile,
	}

	if configFile == "" {
		configFile = os.Getenv(CONFIG_FILE)
	}

	if configFile != "" {
		if err := LoadConfig(configFile, &config); err != nil {
			return nil, err
		}
	}
	return &config, nil
}
