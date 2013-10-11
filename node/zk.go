// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
	zookeeper "github.com/petar/gozk"
	"path"
	"strings"
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
	conn, zch, err := zookeeper.Dial(uri, 5e9)
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

func (zk *ZooKeeper) create(file, info string) (err error) {
	file = path.Clean(file)
	parts := strings.Split(file, "/")
	prefix := "/"
	for i := 0; i < len(parts); i++ {
		prefix = path.Join(prefix, parts[i])
		if _, err = zk.conn.Create(prefix, info, 0,
			zookeeper.WorldACL(zookeeper.PERM_ALL)); err != nil && !isNodeExists(err) {
			return
		}
	}
	return
}

func (zk *ZooKeeper) Register(file string, info []byte) (err error) {
	var stat *zookeeper.Stat
	stat, err = zk.conn.Exists(file)
	if err != nil {
		return
	}
	if stat == nil {
		if err = zk.create(file, string(info)); err != nil {
			return
		}
	} else {
		_, err = zk.conn.Set(file, string(info), stat.Version())
	}
	return
}

func (zk *ZooKeeper) Close() (err error) {
	return zk.conn.Close()
}

func (zk *ZooKeeper) Watch(file string, watcher chan<- []byte) (err error) {
	defer func() {
		if err != nil {
			if zke, ok := err.(*zookeeper.Error); ok {
				if zke.Code == zookeeper.ZCLOSING {
					err = ErrConnection
				}
			}
		}
	}()
	var stat *zookeeper.Stat
	stat, err = zk.conn.Exists(file)
	if err != nil {
		return
	}
	if stat == nil {
		if err = zk.create(file, ""); err != nil {
			return
		}
	} else {
		var w <-chan zookeeper.Event
		if _, stat, w, err = zk.conn.GetW(file); err != nil {
			return err
		}
		ev := <-w
		if ev.Type == zookeeper.EVENT_CHANGED {
			var data string
			if data, _, err = zk.conn.Get(ev.Path); err != nil {
				return
			}
			watcher <- []byte(data)
		}
	}
	return
}

func filterErr(err error) *zookeeper.Error {
	if err == nil {
		return nil
	}
	ze, ok := err.(*zookeeper.Error)
	if !ok {
		return nil
	}
	if ze == nil {
		return nil
	}
	return ze
}

func isNoNode(err error) bool {
	ze := filterErr(err)
	if ze == nil {
		return false
	}
	return ze.Code == zookeeper.ZNONODE
}

func isNodeExists(err error) bool {
	ze := filterErr(err)
	if ze == nil {
		return false
	}
	return ze.Code == zookeeper.ZNODEEXISTS
}
