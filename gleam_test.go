package gleam

import (
	"testing"
	//	"github.com/aarzilli/golua/lua"
	//	"github.com/stevedonovan/luar"
)

func TestGleam(t *testing.T) {
	g := NewGleam("./scripts/")
	if err := g.Init(); err != nil {
		t.Fatal(err)
	}

	if g.config.ClientId != "testing" {
		t.Fatalf("Config error: %+v", g.config)
	}

	if err := g.Final(); err != nil {
		t.Fatal(err)
	}
}
