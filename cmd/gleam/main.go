package main

import (
	"flag"
	"log"

	"github.com/mikespook/gleam"
)

var (
	scripts string
)

func init() {
	flag.StringVar(&scripts, "scripts", "./scripts/", "Lua scripts directory")
	flag.Parse()
}

func main() {
	g := gleam.NewGleam(scripts)
	if err := g.Init(); err != nil {
		log.Fatal(err)
	}

	if err := g.Serve(); err != nil {
		log.Fatal(err)
	}

	if err := g.Final(); err != nil {
		log.Fatal(err)
	}
}
