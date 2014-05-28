package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mikespook/golib/log"
	"gopkg.in/yaml.v1"
)

type Config struct {
	Name   string
	Pid    string
	Script string
	Region []string
	Etcd   []string
	Ca     string
	Cert   string
	Key    string
	Log    struct {
		File  string
		Level string
	}
}

func LoadConfig(filename string, config *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return ParseConfig(data, config)
}

func ParseConfig(data []byte, config *Config) error {
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	return nil
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
	if config.Region == nil {
		config.Region = strings.Split(region, ",")
	}
	if config.Etcd == nil {
		config.Etcd = strings.Split(etcd, ",")
	}
	return &config, nil
}
