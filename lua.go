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

func (luaipt *LuaIpt) Exec(name string, data interface{}) error {
	f := path.Join(luaipt.path, name+".lua")
	luaipt.Bind("Data", luar.NewLuaObjectFromValue(luaipt.state, data))
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

func (gleam *Gleam) luaGet(key string) (string, error) {
	r, err := gleam.client.Get(key, false, false)
	if err != nil {
		return "", err
	}
	return r.Node.Value, nil
}

func (gleam *Gleam) luaGetDir(key string) (map[string]string, error) {
	r, err := gleam.client.Get(key, false, false)
	if err != nil {
		return nil, err
	}
	s := make(map[string]string, r.Node.Nodes.Len())
	for _, n := range r.Node.Nodes {
		s[n.Key] = n.Value
	}
	return s, nil
}

func (gleam *Gleam) luaSet(key, value string, ttl uint64) error {
	_, err := gleam.client.Set(key, value, ttl)
	return err
}

func (gleam *Gleam) luaDelete(key string) error {
	_, err := gleam.client.Delete(key, true)
	return err
}

func (gleam *Gleam) luaWatch(key string) (string, bool, error) {
	r, err := gleam.client.Watch(key, 0, false, nil, nil)
	return r.Node.Value, r.Node.Dir, err
}
