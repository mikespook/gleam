// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    zookeeper "github.com/petar/gozk"
)

type ZooKeeper struct {
    conn *zookeeper.Conn
}

func NewZooKeeper(uri string) (zk *ZooKeeper, err error) {
    zk = new(ZooKeeper)
    err = zk.connect(uri)
    return
}

func (zk *ZooKeeper) connect(uri string) (err error) {
    conn, zch, err := zookeeper.Dial(uri, 3)
    if err != nil {
        return err
    }
    event := <-zch
    if event.State != zookeeper.STATE_CONNECTED {
        conn.Close()
	err = ErrConnection
    }
    zk.conn = conn
    return
}

func (zk *ZooKeeper) Register(file string, info []byte) (err error) {
    return
}

func (zk *ZooKeeper) Close() (err error) {
    return
}

func (zk *ZooKeeper) Watch(file string, watcher chan<- []byte) (err error) {
    return
}
