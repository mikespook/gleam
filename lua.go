package gleam

import (
	"path"

	"github.com/aarzilli/golua/lua"
	"github.com/mikespook/golib/iptpool"
	"github.com/mikespook/golib/log"
	"github.com/stevedonovan/luar"
)

const module = "gleam"

type LuaIpt struct {
	state *lua.State
	path  string
}

func NewLuaIpt() iptpool.ScriptIpt {
	return &LuaIpt{}
}

func (luaipt *LuaIpt) Exec(name string, params interface{}) error {
	f := path.Join(luaipt.path, name+".lua")
	luaipt.Bind("Data", luar.NewLuaObjectFromValue(luaipt.state, params))
	return luaipt.state.DoFile(f)
}

func (luaipt *LuaIpt) Init(path string) error {
	luaipt.state = luar.Init()
	luaipt.Bind("Debugf", log.Debugf)
	luaipt.Bind("Debug", log.Debug)
	luaipt.Bind("Messagef", log.Messagef)
	luaipt.Bind("Message", log.Message)
	luaipt.Bind("Warningf", log.Warningf)
	luaipt.Bind("Warning", log.Warning)
	luaipt.Bind("Errorf", log.Errorf)
	luaipt.Bind("Error", log.Error)
	luaipt.path = path
	return nil
}

func (luaipt *LuaIpt) Final() error {
	luaipt.state.Close()
	return nil
}

func (luaipt *LuaIpt) Bind(name string, item interface{}) error {
	luar.Register(luaipt.state, module, luar.Map{
		name: item,
	})
	return nil
}
