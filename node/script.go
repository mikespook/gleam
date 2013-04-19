// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "os"
    "time"
    "strings"
    "io/ioutil"
)

var (
    defaultPath string
)

func SetDefaultPath(path string) {
    defaultPath = path
}

func nameToPath(name string) string {
    return defaultPath + string(os.PathSeparator) + strings.Replace(name, ".", "", -1)
}

func ifScriptModified(name string, lastmod time.Time, ext string) (c bool, err error) {
    fi, err := os.Stat(nameToPath(name) + "." + ext)
    if err != nil {
        return
    }
    c = !fi.ModTime().Equal(lastmod)
    return
}

func loadScript(name string, ext string) (fi os.FileInfo, contents string, err error) {
    path := nameToPath(name) + "." + ext
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

func isScriptErr(err error) bool {
    return os.IsNotExist(err) || os.IsPermission(err)
}

