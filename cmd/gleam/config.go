package main

import (
	"io/ioutil"

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
