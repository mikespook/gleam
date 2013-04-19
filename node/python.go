// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "os"
    "github.com/qiniu/py"
)

type pyScript struct {
    fi os.FileInfo
    code *py.Code
}

var (
    pyScriptMap map[string]*pyScript
)

func init() {
    pyScriptMap = make(map[string]*pyScript)
}

func Python(name string, params ... interface{}) {
    // check if the script has been loaded
    needLoad := false
    script, ok := pyScriptMap[name]
    if ok {
        // if it was modified
        m, err := ifScriptModified(name, script.fi.ModTime(), "py")
        if isScriptErr(err) {
            pyScriptMap[name].code.Decref()
            delete(pyScriptMap, name)
            _err(ErrLoadScript)
            return
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
            _err(ErrLoadScript)
            return
        }
        code, err := py.Compile(contents, "", py.FileInput)
        if err != nil {
            _err(err)
            return
        }
        pyScriptMap[name] = &pyScript{fi: fi, code: code}
    }

    mod, err := py.ExecCodeModule("z-node", pyScriptMap[name].code.Obj())
    if err != nil {
        _err(err)
    }
    defer mod.Decref()
}
