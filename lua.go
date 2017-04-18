package gleam

import (
	"log"
	"os"
	"path"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	lua "github.com/yuin/gopher-lua"
)

type luaEnv struct {
	root  string
	state *lua.LState
}

func newLuaEnv(root string) *luaEnv {
	return &luaEnv{
		root: root,
	}
}

func (l *luaEnv) Init(config *Config) error {
	opt := lua.Options{
		CallStackSize:       1024,
		IncludeGoStackTrace: true,
		RegistrySize:        1024 * 64,
		SkipOpenLibs:        false,
	}
	l.state = lua.NewState(opt)
	l.state.SetGlobal("Log", l.state.NewFunction(func(L *lua.LState) int {
		log.Print(L.GetTop())
		return 0
	}))
	l.state.SetGlobal("Logf", l.state.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	if err := l.mustDoScript("init"); err != nil {
		return err
	}
	return config.L(l.state)
}

func (l *luaEnv) Final() error {
	err := l.mustDoScript("final")
	l.state.Close()
	return err
}

func (l *luaEnv) mustDoScript(name string) error {
	script := path.Join(l.root, name+".lua")
	if _, err := os.Stat(script); err != nil {
		return err
	}
	if err := l.state.DoFile(script); err != nil {
		return err
	}
	return nil
}

func (l *luaEnv) newHandler(name string) mqtt.MessageHandler {
	script := path.Join(l.root, name+".lua")
	if _, err := os.Stat(script); err != nil {
		return nil
	}
	return func(client mqtt.Client, msg mqtt.Message) {
		state, cancel := l.state.NewThread()
		defer cancel()
		if err := state.DoFile(script); err != nil {
			log.Printf("MsgErr[%X]: %s", msg.MessageID(), err)
		}
		defer state.Close()
	}
}

func (l *luaEnv) defaultHandler(client mqtt.Client, msg mqtt.Message) {

	p := lua.P{
		Fn:      l.state.GetGlobal("DefaultPublishHandler"),
		NRet:    0,
		Protect: true,
	}

	msgL := &lua.LTable{}
	msgL.RawSetString("Duplicate", lua.LBool(msg.Duplicate()))
	msgL.RawSetString("MessageID", lua.LNumber(msg.MessageID()))
	msgL.RawSetString("Payload", lua.LString(msg.Payload()))
	msgL.RawSetString("Qos", lua.LNumber(msg.Qos()))
	msgL.RawSetString("Retained", lua.LBool(msg.Retained()))
	msgL.RawSetString("Topic", lua.LString(msg.Topic()))

	if err := l.state.CallByParam(p, msgL); err != nil {
		log.Printf("DefMsgErr[%s-%X]: %s", msg.Topic(), msg.MessageID(), err)
	}
}
