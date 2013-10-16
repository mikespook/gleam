// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
	"github.com/ha/doozer"
)

type Doozer struct {
	conn     *doozer.Conn
	rev      int64
	infoFile string
	z *ZNode
}

func NewDoozer(uri, buri string) (d *Doozer, err error) {
	d = new(Doozer)
	err = d.connect(uri, buri)
	return
}

func (d *Doozer) connect(uri, buri string) (err error) {
	if d.conn, err = doozer.DialUri(uri, buri); err != nil {
		return
	}
	if d.rev, err = d.conn.Rev(); err != nil {
		return
	}
	return
}

func (d *Doozer) Register(file string, info []byte) (err error) {
	d.infoFile = file
	d.rev, err = d.conn.Set(d.infoFile, d.rev, info)
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
		watcher <- ev.Body
	}
	return
}

func (d *Doozer) SetOnWire(regine, name string, params interface{}) (err error) {
	f := &ZFunc{Name: name, Params: params}
	data, err := d.z.Coder.Encode(f)
	if err != nil {
		return
	}
	d.rev, err = d.conn.Set(MakeWire(regine), d.rev, data)
	if err != nil {
		return
	}
	return
}

func (d *Doozer) SetOnSelf(name string, params interface{}) (err error) {
	f := &ZFunc{Name: name, Params: params}
	data, err := d.z.Coder.Encode(f)
	if err != nil {
		return
	}
	d.rev, err = d.conn.Set(d.z.nodeFile, d.rev, data)
	if err != nil {
		return
	}
	return
}

func (d *Doozer) SetNode(z *ZNode) {
	d.z = z
}
