package gleam

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	// lua
	"github.com/cjoudrey/gluahttp"
	"github.com/cjoudrey/gluaurl"
	"github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
	"layeh.com/gopher-json"
	"layeh.com/gopher-lfs"
	"layeh.com/gopher-luar"

	// mqtt
	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/mikespook/schego"
)

type luaEnv struct {
	sync.RWMutex
	workdir string
	l       *lua.LState
}

func newLuaEnv(workdir string) *luaEnv {
	return &luaEnv{
		workdir: workdir,
	}
}

func (e *luaEnv) Init(config *Config) error {
	if err := os.Chdir(e.workdir); err != nil {
		return err
	}
	e.Lock()
	defer e.Unlock()
	opt := lua.Options{
		CallStackSize:       1024,
		IncludeGoStackTrace: true,
		RegistrySize:        1024 * 64,
		SkipOpenLibs:        false,
	}
	e.l = lua.NewState(opt)
	// Preload module
	json.Preload(e.l)
	lfs.Preload(e.l)
	e.l.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
	e.l.PreloadModule("re", gluare.Loader)
	e.l.PreloadModule("url", gluaurl.Loader)
	// Buildin var & func
	e.setLog()
	e.setLogf()
	e.l.SetGlobal("config", luar.New(e.l, config))
	return e.l.DoFile("bootstrap.lua")
}

func (e *luaEnv) setLog() {
	e.l.SetGlobal("Log", e.l.NewFunction(func(L *lua.LState) int {
		argc := L.GetTop()
		argv := make([]interface{}, argc)
		for i := 1; i <= argc; i++ {
			argv[i-1] = L.Get(i)
		}
		log.Print(argv...)
		return 0
	}))
}

func (e *luaEnv) setLogf() {
	e.l.SetGlobal("Logf", e.l.NewFunction(func(L *lua.LState) int {
		argc := L.GetTop()
		format := L.Get(1).String()
		argv := make([]interface{}, argc-1)
		for i := 2; i <= argc; i++ {
			argv[i-2] = L.Get(i)
		}
		log.Printf(format, argv...)
		return 0
	}))
}

func (e *luaEnv) Final() {
	e.l.Close()
}

func (e *luaEnv) getFn(obj lua.LValue, nest []string) lua.LValue {
	if len(nest) == 0 {
		return lua.LNil
	}
	if obj.Type() == lua.LTTable {
		obj = obj.(*lua.LTable).RawGetString(nest[0])
		e.getFn(obj, nest[1:])
	}
	if obj.Type() == lua.LTFunction {
		return obj
	}
	return lua.LNil
}

func (e *luaEnv) GetFn(name string) lua.LValue {
	nest := strings.Split(name, ".")
	obj := e.l.GetGlobal(nest[0])
	return e.getFn(obj, nest[1:])
}

func (e *luaEnv) newMQTTFunc(name string) mqtt.MessageHandler {
	if name == "" {
		return nil
	}
	e.RLock()
	p := lua.P{
		Fn:      e.GetFn(name),
		Protect: true,
	}
	e.RUnlock()
	if p.Fn.Type() == lua.LTNil { // Final is not defined, return nil to run the default one
		return nil
	}
	return func(client mqtt.Client, msg mqtt.Message) {
		e.Lock()
		L, cancel := e.l.NewThread()
		defer func() {
			if cancel != nil {
				cancel()
			}
			L.Close()
		}()
		clientL := luar.New(e.l, client)
		msgL := messageToLua(e.l, msg)
		e.Unlock()
		if err := L.CallByParam(p, clientL, msgL); err != nil {
			// TODO using lua error func
			log.Printf("Error[%s-%X]: %s", name, msg.MessageID(), err)
		}
	}
}

func (e *luaEnv) newScheduleFunc(name string, client mqtt.Client) schego.ExecFunc {
	e.RLock()
	p := lua.P{
		Fn:      e.GetFn(name),
		Protect: true,
	}
	e.RUnlock()
	if p.Fn.Type() == lua.LTNil { // Specific schego func is not defined, return nil to skip it
		return nil
	}
	return func(ctx context.Context) error {
		L, cancel := e.l.NewThread()
		defer func() {
			if cancel != nil {
				cancel()
			}
			L.Close()
		}()
		clientL := luar.New(L, client)
		return L.CallByParam(p, clientL)
	}
}

func (e *luaEnv) errorFunc(ctx context.Context, err error) {
	e.RLock()
	p := lua.P{
		Fn:      e.l.GetGlobal("onError"),
		NRet:    0,
		Protect: true,
	}
	e.RUnlock()
	if p.Fn.Type() == lua.LTNil {
		return
	}

	errL := luar.New(e.l, err)

	if err := e.l.CallByParam(p, errL); err != nil {
		log.Printf("Scheduler Error: %s", err)
	}
}

func (e *luaEnv) onEvent(name string, client mqtt.Client) error {
	e.RLock()
	p := lua.P{
		Fn:      e.l.GetGlobal(name),
		NRet:    0,
		Protect: true,
	}
	e.RUnlock()
	if p.Fn.Type() == lua.LTNil {
		return nil
	}
	clientL := luar.New(e.l, client)
	return e.l.CallByParam(p, clientL)
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
