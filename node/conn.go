// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

type Conn interface {
    Register(file string, info []byte) error
    Close() error
    Watch(file string, watcher chan<- []byte) error
}
