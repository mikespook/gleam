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

const (
	MessageFunc        = "onDefaultMessage"
	ErrorFunc          = "onError"
	AfterInitFunc      = "afterInit"
	BeforeFinalizeFunc = "beforeFinalize"

	LogFunc  = "log"
	LogfFunc = "logf"

	ConfigVar = "config"

	LuaCallStackSize       = 1024
	LuaIncludeGoStackTrace = true
	LuaRegistrySize        = 1024 * 64
	LuaSkipOpenLibs        = false

	BootstrapFile = "bootstrap.lua"
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
		CallStackSize:       LuaCallStackSize,
		IncludeGoStackTrace: LuaIncludeGoStackTrace,
		RegistrySize:        LuaRegistrySize,
		SkipOpenLibs:        LuaSkipOpenLibs,
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
	e.l.SetGlobal(ConfigVar, luar.New(e.l, config))
	return e.l.DoFile(BootstrapFile)
}

func (e *luaEnv) setLog() {
	e.l.SetGlobal(LogFunc, e.l.NewFunction(func(L *lua.LState) int {
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
	e.l.SetGlobal(LogfFunc, e.l.NewFunction(func(L *lua.LState) int {
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

func (e *luaEnv) getFuncByName(obj lua.LValue, nest []string) lua.LValue {
	if obj.Type() == lua.LTFunction {
		return obj
	}
	if obj.Type() == lua.LTTable {
		if len(nest) == 0 {
			return lua.LNil
		}
		obj = obj.(*lua.LTable).RawGetString(nest[0])
		e.getFuncByName(obj, nest[1:])
	}
	return lua.LNil
}

func (e *luaEnv) GetFuncByName(name string) lua.LValue {
	nest := strings.Split(name, ".")
	obj := e.l.GetGlobal(nest[0])
	return e.getFuncByName(obj, nest[1:])
}

func (e *luaEnv) newOnMessage(name string) mqtt.MessageHandler {
	if name == "" {
		return nil
	}
	e.RLock()
	p := lua.P{
		Fn:      e.GetFuncByName(name),
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
			ctx := context.Background()
			ctx = context.WithValue(ctx, "name", name)
			ctx = context.WithValue(ctx, "message", msg)
			e.onError(ctx, err)
		}
	}
}

func (e *luaEnv) newOnSchedule(name string, client mqtt.Client) schego.ExecFunc {
	e.RLock()
	p := lua.P{
		Fn:      e.GetFuncByName(name),
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
		ctxL := contextToLua(e.l, ctx)
		clientL := luar.New(L, client)
		return L.CallByParam(p, ctxL, clientL)
	}
}

func (e *luaEnv) onError(ctx context.Context, err error) {
	e.RLock()
	p := lua.P{
		Fn:      e.l.GetGlobal(ErrorFunc),
		NRet:    0,
		Protect: true,
	}
	e.RUnlock()
	if p.Fn.Type() == lua.LTNil {
		return
	}

	ctxL := contextToLua(e.l, ctx)
	errL := luar.New(e.l, err.Error())

	if err := e.l.CallByParam(p, ctxL, errL); err != nil {
		log.Printf("Error: %s", err)
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

func contextToLua(L *lua.LState, ctx context.Context) lua.LValue {
	name := ctx.Value("name")
	if name != nil {
		strname, ok := name.(string)
		if !ok {
			return lua.LNil
		}
		msgL := luar.New(L, ctx.Value("message"))
		ctxL := L.CreateTable(0, 2)
		ctxL.RawSetString("Id", lua.LString(strname))
		ctxL.RawSetString("Message", msgL)
		return ctxL
	}
	event := ctx.Value("event")
	if event != nil {
		return luar.New(L, event)
	}
	return lua.LNil
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
