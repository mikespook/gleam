package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

type ConfigEtcd struct {
	Url  string
	Ca   string
	Cert string
	Key  string
}

type Config struct {
	Name   string
	Pid    string
	Script string
	Region []string
	Etcd   ConfigEtcd
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(data)
}

func ParseConfig(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
