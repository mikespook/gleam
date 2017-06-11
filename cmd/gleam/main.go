package main

import (
	"flag"
	"log"

	"github.com/mikespook/gleam"
)

var (
	scripts string
	version string
	showVer bool
)

func init() {
	flag.StringVar(&scripts, "scripts", "./scripts/", "Lua scripts directory")
	flag.BoolVar(&showVer, "version", false, "Show version")
	flag.Parse()
}

func main() {
	if showVer {
		log.Printf("Version: %s", version)
		return
	}
	g := gleam.NewGleam(scripts)
	if err := g.Init(); err != nil {
		log.Fatal(err)
	}

	g.Serve()

	if err := g.Final(); err != nil {
		log.Fatal(err)
	}
}
