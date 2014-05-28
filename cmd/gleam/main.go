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
	if config.Ca != "" || config.Cert != "" || config.Key != "" {
		log.Messagef("Setting TLS (CA=%s; Cert=%s; Key=%s)...", config.Ca, config.Cert, config.Key)
	}
	g, err := gleam.New(config.Etcd, config.Script, config.Cert, config.Key, config.Ca)
	if err != nil {
		log.Error(err)
		return
	}
	defer g.Close()
	g.ErrHandler = func(err error) {
		log.Error(err)
	}
	log.Messagef("Watching(Name = %s)...", config.Name)
	g.WatchNode(config.Name)
	for _, r := range config.Region {
		log.Messagef("Watching(Region = %s)...", r)
		g.WatchRegion(r)
	}
	go g.Serve()
	// signal handler
	sh := signal.NewHandler()
	sh.Bind(os.Interrupt, func() bool { return true })
	go func() {
		g.Wait()
		if err := signal.Send(os.Getpid(), os.Interrupt); err != nil {
			panic(err)
		}
	}()

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
		flag.StringVar(&etcd, "etcd", "http://127.0.0.1:4001", "A comma-delimited list of etcd")
		flag.StringVar(&caFile, "ca-file", "", "Path to the CA file")
		flag.StringVar(&certFile, "cert-file", "", "Path to the cert file")
		flag.StringVar(&keyFile, "key-file", "", "Path to the key file")
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
		Log: struct {
			File  string
			Level string
		}{*log.LogFile, *log.LogLevel},
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
