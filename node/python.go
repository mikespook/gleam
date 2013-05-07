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
    code string
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
    pyScriptMap = make(map[string]*pyScript)
    zNodeMod = new(Module)
)

func PyExec(name string, params ... interface{}) error {
    // check if the script has been loaded
    needLoad := false
    script, ok := pyScriptMap[name]
    if ok {
        // if it was modified
        m, err := ifScriptModified(name, script.fi.ModTime(), "py")
        if isScriptErr(err) {
            delete(pyScriptMap, name)
            return ErrLoadScript
        }
        if m {
            needLoad = true
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
        pyScriptMap[name] = &pyScript{fi: fi, code: contents}
    }

    py.Initialize()
    // znode module
    pyMod, err := py.NewGoModule("znode", "", zNodeMod)
    if err != nil {
        return err
    }
    defer pyMod.Decref()

    d := py.NewDict()
    if err := d.SetItemString("__builtins__", py.GetBuiltins());
        err != nil {
        return err
    }
    defer d.Decref()

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

    code, err := py.Compile(pyScriptMap[name].code, name, py.FileInput)
    if err != nil {
        return err
    }
    defer code.Decref()

    if err := code.Run(d.Obj(), locals.Obj()); err != nil {
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
