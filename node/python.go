// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "os"
    "time"
    "reflect"
    "io/ioutil"
    "github.com/qiniu/py"
    "github.com/mikespook/golib/log"
)

type module struct {}

func (m *module) paserArgs(args *py.Tuple) string {
    var msg string
    for i := 0; i < args.Size(); i ++ {
        if item, err := args.GetItem(i); err == nil {
            msg += item.String()
        }
    }
    return msg
}

func (m *module) Py_debug(args *py.Tuple) (ret *py.Base, err error) {
    log.Debug(m.paserArgs(args))
    return py.IncNone(), nil
}

func (m *module) Py_msg(args *py.Tuple) (ret *py.Base, err error) {
    log.Message(m.paserArgs(args))
    return py.IncNone(), nil
}

func (m *module) Py_warning(args *py.Tuple) (ret *py.Base, err error) {
    log.Warning(m.paserArgs(args))
    return py.IncNone(), nil
}

func (m *module) Py_error(args *py.Tuple) (ret *py.Base, err error) {
    log.Errorf("%s", m.paserArgs(args))
    return py.IncNone(), nil
}
// -----------------------------
type pyCodeInfo struct {
    fi os.FileInfo
    code string
}

type Py struct{
    scripts map[string]*pyCodeInfo
    zNodeMod py.GoModule
    path string
}

func (p *Py) Init(path string) (err error) {
    p.path = path
    p.scripts = make(map[string]*pyCodeInfo)
    py.Initialize()
    p.zNodeMod, err = py.NewGoModule("znode", "", new(module))
    return
}

func (p *Py) Final() (err error) {
    p.zNodeMod.Decref()
    return
}

func (p *Py) Exec(name string, params ... interface{}) error {
    // check if the script has been loaded
    needLoad := false
    script, ok := p.scripts[name]
    if ok {
        // if it was modified
        m, err := ifPyFileModified(p.path, name, script.fi.ModTime(), "py")
        if isPyFileErr(err) {
            delete(p.scripts, name)
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
        fi, contents, err := loadPyFile(p.path, name, "py")
        if isPyFileErr(err) {
            return ErrLoadScript
        }
        p.scripts[name] = &pyCodeInfo{fi: fi, code: contents}
    }

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
    if err := locals.SetItemString("_ARGS_", args.Obj()); err != nil {
        return err
    }
    if err := locals.SetItemString("_ROOT_", py.NewString(p.path).Obj()); err != nil {
        return err
    }
    defer locals.Decref()

    code, err := py.Compile(p.scripts[name].code, name, py.FileInput)
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
    case reflect.Float32, reflect.Float64:
        item = py.NewFloat(r.Float()).Obj()
    default:
        item = py.IncNone()
    }
    return item
}

func ifPyFileModified(path, name string, lastmod time.Time, ext string) (c bool, err error) {
    fi, err := os.Stat(nameToPath(path, name) + "." + ext)
    if err != nil {
        return
    }
    c = !fi.ModTime().Equal(lastmod)
    return
}

func loadPyFile(path, name string, ext string) (fi os.FileInfo, contents string, err error) {
    path = nameToPath(path, name) + "." + ext
    fi, err = os.Stat(path)
    if err != nil {
        return
    }
    c, err := ioutil.ReadFile(path)
    if err != nil {
        return
    }
    contents = string(c)
    return
}

func isPyFileErr(err error) bool {
    return os.IsNotExist(err) || os.IsPermission(err)
}
