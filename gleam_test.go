package gleam

import (
	"os"
	"strings"
	"testing"
)

var (
	wd string
)

func init() {
	var err error
	if wd, err = os.Getwd(); err != nil {
		panic(err)
	}
}

func resetWD() {
	if err := os.Chdir(wd); err != nil {
		panic(err)
	}
}

func TestGleam(t *testing.T) {
	resetWD()

	g := NewGleam("./scripts/")
	if err := g.Init(); err != nil {
		t.Fatal(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(g.config.ClientId, hostname) {
		t.Fatalf("Config error: %+v", g.config)
	}

	if err := g.Final(); err != nil {
		t.Fatal(err)
	}
}
