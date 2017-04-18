package gleam

import (
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type Config struct {
	Brokers     []string
	Prefix      string
	ClientId    string
	StateUpdate int
	Tasks       map[string]byte
}

func (config *Config) L(state *lua.LState) error {
	return gluamapper.Map(state.GetGlobal("config").(*lua.LTable), &config)
}
