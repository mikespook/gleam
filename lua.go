package gleam

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/cjoudrey/gluahttp"
	"github.com/cjoudrey/gluaurl"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mikespook/schego"
	"github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
	"layeh.com/gopher-json"
	"layeh.com/gopher-lfs"
	"layeh.com/gopher-luar"
)

type luaEnv struct {
	sync.RWMutex
	root  string
	state *lua.LState
}

func newLuaEnv(root string) *luaEnv {
	return &luaEnv{
		root: root,
	}
}

func (l *luaEnv) Init(config *Config) error {
	l.Lock()
	defer l.Unlock()
	if err := os.Chdir(l.root); err != nil {
		return err
	}
	opt := lua.Options{
		CallStackSize:       1024,
		IncludeGoStackTrace: true,
		RegistrySize:        1024 * 64,
		SkipOpenLibs:        false,
	}
	l.state = lua.NewState(opt)
	json.Preload(l.state)
	lfs.Preload(l.state)
	l.state.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
	l.state.PreloadModule("re", gluare.Loader)
	l.state.PreloadModule("url", gluaurl.Loader)
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
	script := name + ".lua"
	if _, err := os.Stat(script); err != nil {
		return err
	}
	if err := l.state.DoFile(script); err != nil {
		return err
	}
	return nil
}

func (l *luaEnv) newMQTTHandler(name string) mqtt.MessageHandler {
	script := name + ".lua"
	if _, err := os.Stat(script); err != nil {
		return nil
	}
	return func(client mqtt.Client, msg mqtt.Message) {
		l.Lock()
		state, cancel := l.state.NewThread()
		defer func() {
			if cancel != nil {
				cancel()
			}
			state.Close()
		}()
		clientL := luar.New(l.state, client)
		state.SetGlobal("Client", clientL)
		msgL := messageToLua(l.state, msg)
		state.SetGlobal("Message", msgL)
		l.Unlock()
		if err := state.DoFile(script); err != nil {
			log.Printf("Error[%s-%X]: %s", name, msg.MessageID(), err)
		}
	}
}

func (l *luaEnv) newExecFunc(name string, client mqtt.Client) schego.ExecFunc {
	script := name + ".lua"
	if _, err := os.Stat(script); err != nil {
		return func(ctx context.Context) error {
			l.RLock()
			p := lua.P{
				Fn:      l.state.GetGlobal("ScheduleDefaultFunc"),
				NRet:    0,
				Protect: true,
			}
			l.RUnlock()
			if p.Fn.Type() == lua.LTNil {
				return nil
			}
			ctxL := luar.New(l.state, ctx)
			clientL := luar.New(l.state, client)
			return l.state.CallByParam(p, ctxL, clientL)
		}

	}
	return func(ctx context.Context) error {
		l.Lock()
		state, cancel := l.state.NewThread()
		defer func() {
			if cancel != nil {
				cancel()
			}
			state.Close()
		}()
		clientL := luar.New(l.state, client)
		state.SetGlobal("Client", clientL)
		l.Unlock()
		return state.DoFile(script)
	}
}

func (l *luaEnv) errorFunc(ctx context.Context, err error) {
	l.RLock()
	p := lua.P{
		Fn:      l.state.GetGlobal("ErrorFunc"),
		NRet:    0,
		Protect: true,
	}
	l.RUnlock()
	if p.Fn.Type() == lua.LTNil {
		return
	}

	ctxL := luar.New(l.state, ctx)
	errL := luar.New(l.state, err)

	if err := l.state.CallByParam(p, ctxL, errL); err != nil {
		log.Printf("Scheduler Error: %s", err)
	}
}

func (l *luaEnv) defaultMQTTHandler(client mqtt.Client, msg mqtt.Message) {
	l.RLock()
	p := lua.P{
		Fn:      l.state.GetGlobal("MQTTDefaultHandler"),
		NRet:    0,
		Protect: true,
	}
	l.RUnlock()
	if p.Fn.Type() == lua.LTNil {
		return
	}

	clientL := luar.New(l.state, client)
	msgL := messageToLua(l.state, msg)

	if err := l.state.CallByParam(p, clientL, msgL); err != nil {
		log.Printf("Error[%s-%X]: %s", msg.Topic(), msg.MessageID(), err)
	}
}

func (l *luaEnv) onEvent(name string, client mqtt.Client) error {
	l.RLock()
	p := lua.P{
		Fn:      l.state.GetGlobal(name),
		NRet:    0,
		Protect: true,
	}
	l.RUnlock()
	if p.Fn.Type() == lua.LTNil {
		return nil
	}
	clientL := luar.New(l.state, client)
	return l.state.CallByParam(p, clientL)
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
