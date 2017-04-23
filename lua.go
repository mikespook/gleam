package gleam

import (
	"log"
	"os"
	"path"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
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
		argc := L.GetTop()
		argv := make([]interface{}, argc)
		for i := 1; i <= argc; i++ {
			argv[i-1] = L.Get(i)
		}
		log.Print(argv...)
		return 0
	}))
	l.state.SetGlobal("Logf", l.state.NewFunction(func(L *lua.LState) int {
		argc := L.GetTop()
		format := L.Get(1).String()
		argv := make([]interface{}, argc-1)
		for i := 2; i <= argc; i++ {
			argv[i-2] = L.Get(i)
		}
		log.Printf(format, argv...)
		return 0
	}))
	l.state.SetGlobal("config", luar.New(l.state, config))
	return l.mustDoScript("init")
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
		defer func() {
			if cancel != nil {
				cancel()
			}
		}()

		clientL := luar.New(l.state, client)
		state.SetGlobal("Client", clientL)
		msgL := messageToLua(l.state, msg)
		state.SetGlobal("Message", msgL)

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

	clientL := luar.New(l.state, client)
	msgL := messageToLua(l.state, msg)

	if err := l.state.CallByParam(p, clientL, msgL); err != nil {
		log.Printf("DefMsgErr[%s-%X]: %s", msg.Topic(), msg.MessageID(), err)
	}
}

func messageToLua(L *lua.LState, msg mqtt.Message) *lua.LTable {
	msgL := L.CreateTable(0, 6)
	msgL.RawSetString("Duplicate", lua.LBool(msg.Duplicate()))
	msgL.RawSetString("MessageID", lua.LNumber(msg.MessageID()))
	msgL.RawSetString("Payload", lua.LString(msg.Payload()))
	msgL.RawSetString("Qos", lua.LNumber(msg.Qos()))
	msgL.RawSetString("Retained", lua.LBool(msg.Retained()))
	msgL.RawSetString("Topic", lua.LString(msg.Topic()))
	return msgL
}
