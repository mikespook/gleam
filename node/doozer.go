// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "github.com/ha/doozer"
)

type Doozer struct {
    conn *doozer.Conn
    uri, buri string
    rev int64
    infoFile string
}

func NewDoozer(uri, buri string) (d *Doozer, err error) {
    d = new(Doozer)
    err = d.Connect(uri, buri)
    return
}

func (d *Doozer) Connect(params ... string) (err error) {
    if len(params) != 2 {
        return ErrParam
    }
    d.uri = params[0]
    d.buri = params[1]

    if d.conn, err = doozer.DialUri(d.uri, d.buri); err != nil {
        return
    }
    if d.rev, err = d.conn.Rev(); err != nil {
        return
    }
    return
}

func (d *Doozer) Register(file string, info []byte) (err error) {
    d.infoFile = file
    d.rev, err = d.conn.Set(file, d.rev, info)
    return
}

func (d *Doozer) Close() (err error) {
    if err = d.conn.Del(d.infoFile, d.rev); err != nil {
        return
    }
    d.conn.Close()
    return
}

func (d *Doozer) Watch(file string, watcher chan<- []byte) (err error) {
    ev, err := d.conn.Wait(file, d.rev)
    if err != nil {
        if err == doozer.ErrClosed {
            err = ErrConnection
            return
        }
        return
    }
    d.rev = ev.Rev + 1
    if ev.IsSet() {
        watcher <-ev.Body
    }
    return
}
