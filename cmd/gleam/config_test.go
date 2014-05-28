package main

import (
	"testing"
)

var data = `
name: test-node
pid: /tmp/test.pid
script: /tmp/gleam
region: [abc, def, ghi]
etcd: ["127.0.0.1:4001", "127.0.0.1:4002"]
ca:
cert:
key:
log:
 file: /tmp/gleam.log
 level: all
`

func TestParseConfig(t *testing.T) {
	var config Config
	if err := ParseConfig([]byte(data), &config); err != nil {
		t.Error(err)
	}
	if config.Name != "test-node" {
		t.Errorf("Wrong name: %s", config.Name)
	}
	if config.Pid != "/tmp/test.pid" {
		t.Errorf("Wrong pid: %s", config.Pid)
	}
	if config.Script != "/tmp/gleam" {
		t.Errorf("Wrong script: %s", config.Script)
	}
	if len(config.Region) != 3 && config.Region[2] != "ghi" {
		t.Errorf("Wrong region: %#v", config.Region)
	}
	if len(config.Etcd) != 2 && config.Etcd[0] != "127.0.0.1:4001" {
		t.Errorf("Wrong etcd url: %#v", config.Etcd)
	}
	if config.Ca != "" {
		t.Errorf("Wrong etcd ca: %s", config.Ca)
	}
	if config.Cert != "" {
		t.Errorf("Wrong etcd ca: %s", config.Cert)
	}
	if config.Key != "" {
		t.Errorf("Wrong etcd ca: %s", config.Key)
	}
	if config.Log.File != "/tmp/gleam.log" {
		t.Errorf("Wrong log file: %s", config.Log.File)
	}
	if config.Log.Level != "all" {
		t.Errorf("Wrong log level: %s", config.Log.Level)
	}
}
