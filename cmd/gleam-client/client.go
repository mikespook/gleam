package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mikespook/gleam"
)

func main() {
	// prepare the configuration
	config := InitConfig()
	if config == nil {
		return
	}
	client, err := gleam.NewClient(config.Etcd, config.Cert, config.Key, config.Ca)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch config.Cmd {
	case "info":
		fmt.Println("Info:")
		m, err := client.Info()
		if err != nil {
			fmt.Println(err)
			return
		}
		if len(m) == 0 {
			fmt.Println("\tNone")
			return
		}
		for k, v := range m {
			fmt.Printf("\t%s =>\n", k)
			fmt.Printf("\t\tcreated\t=> %s\n", v[k+"/created"])
			fmt.Printf("\t\tlast\t=> %s\n", v[k+"/last"])
			fmt.Printf("\t\terror\t=> %s\n", v[k+"/error"])
		}
	case "region":
		fmt.Println("Regions:")
		m, err := client.List(gleam.RegionDir)
		if err != nil {
			fmt.Printf("\t%s\n", err)
			return
		}
		for k, v := range m {
			fmt.Printf("\t%s => %s\n", k, v)
		}
	case "node":
		fmt.Println("Node:")
		m, err := client.List(gleam.NodeDir)
		if err != nil {
			fmt.Printf("\t%s\n", err)
			return
		}
		for k, v := range m {
			fmt.Printf("\t%s => %s\n", k, v)
		}
	case "call":
		if err := client.Call(flag.Arg(1), flag.Arg(2), flag.Arg(3)); err != nil {
			fmt.Println(err)
			return
		}
	default:
		flag.Usage()
	}
}

type Config struct {
	Etcd []string
	Ca   string
	Cert string
	Key  string

	Cmd string
}

func InitConfig() *Config {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [Options] [Command]:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nCommand includes :")
		fmt.Fprintf(os.Stderr, "\t call: Call a function on nodes file\n")
		fmt.Fprintf(os.Stderr, "\t region: List all regions\n")
		fmt.Fprintf(os.Stderr, "\t node: List all nodes\n")
		fmt.Fprintf(os.Stderr, "\t info: List all nodes info\n")
	}
	var etcd, caFile, certFile, keyFile string
	if !flag.Parsed() {
		flag.StringVar(&etcd, "etcd", "http://127.0.0.1:4001", "A comma-delimited list of etcd")
		flag.StringVar(&caFile, "ca-file", "", "Path to the CA file")
		flag.StringVar(&certFile, "cert-file", "", "Path to the cert file")
		flag.StringVar(&keyFile, "key-file", "", "Path to the key file")

		flag.Parse()
	}

	return &Config{
		Etcd: strings.Split(etcd, ","),
		Ca:   caFile,
		Cert: certFile,
		Key:  keyFile,
		Cmd:  flag.Arg(0),
	}
}
