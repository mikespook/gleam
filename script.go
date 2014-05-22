// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
	"os"
	"strings"
)

type ScriptInterpreter interface {
	Exec(name string, params ...interface{}) error
	Init(path string) error
	Final() error
}

func nameToPath(path, name string) string {
	return path + string(os.PathSeparator) + strings.Replace(name, ".", "", -1)
}
