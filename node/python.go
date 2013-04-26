// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "os"
    "reflect"
    "github.com/qiniu/py"
    "github.com/mikespook/golib/log"
)

type pyScript struct {
    fi os.FileInfo
    code *py.Code
}

type Module struct {}

func (m *Module) paserArgs(args *py.Tuple) string {
    var msg string
    for i := 0; i < args.Size(); i ++ {
        if item, err := args.GetItem(i); err == nil {
            msg += item.String()
        }
    }
    return msg
}

func (m *Module) Py_debug(args *py.Tuple) (ret *py.Base, err error) {
    log.Debug(m.paserArgs(args))
    return py.IncNone(), nil
}

func (m *Module) Py_msg(args *py.Tuple) (ret *py.Base, err error) {
    log.Message(m.paserArgs(args))
    return py.IncNone(), nil
}

func (m *Module) Py_warning(args *py.Tuple) (ret *py.Base, err error) {
    log.Warning(m.paserArgs(args))
    return py.IncNone(), nil
}

func (m *Module) Py_error(args *py.Tuple) (ret *py.Base, err error) {
    log.Errorf("%s", m.paserArgs(args))
    return py.IncNone(), nil
}

var (
    pyScriptMap map[string]*pyScript
    zNodeMod py.GoModule
    zDict *py.Base
)

func PyInit() (err error) {
    py.Initialize()
    pyScriptMap = make(map[string]*pyScript)
    zNodeMod, err = py.NewGoModule("znode", "", new(Module))
    if err != nil {
        return
    }

    d := py.NewDict()
    if err = d.SetItemString("__builtins__", py.GetBuiltins());
        err != nil {
        return
    }
    zDict = d.Obj()
    return
}

func PyClose() {
    zNodeMod.Decref()
    zDict.Decref()
    py.Finalize()
}

func PyExec(name string, params ... interface{}) error {
    // check if the script has been loaded
    needLoad := false
    script, ok := pyScriptMap[name]
    if ok {
        // if it was modified
        m, err := ifScriptModified(name, script.fi.ModTime(), "py")
        if isScriptErr(err) {
            pyScriptMap[name].code.Decref()
            delete(pyScriptMap, name)
            return ErrLoadScript
        }
        if m {
            needLoad = true
            pyScriptMap[name].code.Decref()
        }
    } else {
        needLoad = true
    }

    // load the script
    if needLoad {
        fi, contents, err := loadScript(name, "py")
        if isScriptErr(err) {
            return ErrLoadScript
        }
        code, err := py.Compile(contents, name, py.FileInput)
        if err != nil {
            return err
        }
        pyScriptMap[name] = &pyScript{fi: fi, code: code}
    }

    // init module args
    args := py.NewTuple(len(params))
    for i, v := range params {
        if err := args.SetItem(i, parsePyArgs(v)); err != nil {
            return err
        }
    }
    defer args.Decref()

    locals := py.NewDict()
    if err := locals.SetItemString("args", args.Obj()); err != nil {
        return err
    }
    defer locals.Decref()

    if err := pyScriptMap[name].code.Run(zDict, locals.Obj()); err != nil {
        return err
    }
    return nil
}

func parsePyArgs(arg interface{}) (item *py.Base) {
    r := reflect.ValueOf(arg)
    switch r.Kind() {
    case reflect.Bool, reflect.Int, reflect.Int16, reflect.Int32,
        reflect.Uint, reflect.Uint16, reflect.Uint32,
        reflect.Int64, reflect.Uint64:
        item = py.NewInt64(r.Int()).Obj()
    case reflect.String:
        item = py.NewString(r.String()).Obj()
    case reflect.Array, reflect.Slice:
        t := py.NewTuple(r.Len())
        for i := 0; i < r.Len(); i ++ {
            if err := t.SetItem(i, parsePyArgs(r.Index(i).Interface()));
                err != nil {
                return py.IncNone()
            }
        }
        item = t.Obj()
    case reflect.Map:
        d := py.NewDict()
        for _, v := range r.MapKeys() {
            key := parsePyArgs(v.Interface())
            value := parsePyArgs(r.MapIndex(v).Interface())
            if err := d.SetItem(key, value); err != nil {
                return py.IncNone()
            }
        }
        item = d.Obj()
    default:
        item = py.IncNone()
    }
    return item
}
