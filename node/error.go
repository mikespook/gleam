// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "errors"
)

var (
    ErrHandler ErrorHandlerFunc
    ErrLoadScript = errors.New("Error loading script")
)

type ErrorHandlerFunc func(error)


func _err(err error) {
    if ErrHandler != nil {
        ErrHandler(err)
    }
}
