package main

import (
	"testing"
)

var data = `
name: test-node
pid: /tmp/test.pid
script: /tmp/gleam
region: [abc, def, ghi]
etcd:
 url: 127.0.0.1:7001
 ca:
 cert:
 key:
`

func TestParseConfig(t *testing.T) {
	config, err := ParseConfig([]byte(data))
	if err != nil {
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
		t.Errorf("Wrong region: %v", config.Region)
	}
	if config.Etcd.Url != "127.0.0.1:7001" {
		t.Errorf("Wrong etcd url: %s", config.Etcd.Url)
	}
	if config.Etcd.Ca != "" {
		t.Errorf("Wrong etcd ca: %s", config.Etcd.Ca)
	}
	if config.Etcd.Cert != "" {
		t.Errorf("Wrong etcd ca: %s", config.Etcd.Cert)
	}
	if config.Etcd.Key != "" {
		t.Errorf("Wrong etcd ca: %s", config.Etcd.Key)
	}
}
