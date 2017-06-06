package gleam

import (
	"testing"
	//	"github.com/aarzilli/golua/lua"
	//	"github.com/stevedonovan/luar"
)

func TestLuaEnv(t *testing.T) {
	resetWD()

	e := newLuaEnv("./scripts/")
	var config Config
	if err := e.Init(&config); err != nil {
		t.Fatal(err)
	}
	e.Final()
}
