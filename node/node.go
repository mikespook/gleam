// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package node

import (
    "os"
    "fmt"
    "sync"
    "time"
    "encoding/json"
    "github.com/ha/doozer"
    "github.com/mikespook/golib/funcmap"
    "github.com/mikespook/golib/log"
)

const (
    MaxErrorCount = 5
    FuncPoolSize = 30
    DefaultRegion = "default"
    Root = "/z-node"
    WireFile = Root + "/%s/wire"
    NodeFile = Root + "/node/%s/%d"
    InfoFile = Root + "/info/%s/%d"
)

type ZNode struct {
    ErrHandler ErrorHandlerFunc

    conn *doozer.Conn
    uri, buri, hostname string
    rev int64

    wires []string
    nodeFile, infoFile string
    fmap funcmap.Funcs
    w sync.WaitGroup
}

type ZFunc struct {
    Name string
    Params []interface{}
}

func MakeWire(region string) string {
    return fmt.Sprintf(WireFile, region)
}

func MakeNode(file, hostname string, pid int) string {
    return fmt.Sprintf(file, hostname, pid)
}

func New(hostname string, regions ... string) (node *ZNode) {
    if len(regions) == 0 {
        regions = []string{DefaultRegion}
    }
    for i, v := range regions {
        regions[i] = MakeWire(v)
    }
    pid := os.Getpid()
    nodeFile := MakeNode(NodeFile, hostname, pid)
    infoFile := MakeNode(InfoFile, hostname, pid)
    return &ZNode {
        wires: regions,
        nodeFile: nodeFile,
        infoFile: infoFile,
        fmap: funcmap.New(FuncPoolSize),
    }
}

func (node *ZNode) Bind(name string, fn interface{}) (err error) {
    return node.fmap.Bind(name, fn)
}

func (node *ZNode) Start(uri, buri string) (err error) {
    node.uri = uri
    node.buri = buri

    if node.conn, err = doozer.DialUri(node.uri, node.buri); err != nil {
        return
    }
    if node.rev, err = node.conn.Rev(); err != nil {
        return
    }
    if node.rev, err = node.conn.Set(
        node.infoFile,
        node.rev, []byte(time.Now().String()));
        err != nil {
        return
    }
    node.watchSelf()
    node.watchWire()
    return
}

func (node *ZNode) Close() {
    if err := node.conn.Del(
        node.infoFile,
        node.rev); err != nil {
        node.err(err)
    }
    if err := node.conn.Del(
        node.nodeFile,
        node.rev); err != nil {
        node.err(err)
    }
    node.conn.Close()
    zNodeMod.Decref()
}

func (node *ZNode) Wait() {
    node.w.Wait()
}

func (node *ZNode) err(err error) {
    if node.ErrHandler != nil {
        node.ErrHandler(err)
    }
}

func (node *ZNode) watch(file string) {
    defer node.w.Done()
    for i := 0; i < MaxErrorCount; i ++ {
        ev, err := node.conn.Wait(file, node.rev)
        if err != nil {
            if err == doozer.ErrClosed {
                break
            }
            node.err(err)
            continue
        }
        node.rev = ev.Rev + 1
        if ev.IsSet() {
            var fn ZFunc
            if err := json.Unmarshal(ev.Body, &fn); err != nil {
                node.err(err)
                continue
            }
            go node.Call(fn.Name, fn.Params ...)
            i = 0
        }
    }
}

func (node *ZNode) Call(name string, params ... interface{}) {
    if _, ok := node.fmap[name]; ok {
        log.Messagef("Call Go function %s, %t supplied.", name, params)
        if _, err := node.fmap.Call(name, params ...); err != nil {
            node.err(err)
        }
        return
    }
    log.Messagef("Call Python script %s, %t supplied.", name, params)
    if err := execPython(name, params ...); err != nil {
        node.err(err)
    }
}

func (node *ZNode) watchSelf() {
    node.w.Add(1)
    go node.watch(node.nodeFile)
}

func (node *ZNode) watchWire() {
    for _, v := range node.wires {
        node.w.Add(1)
        go node.watch(v)
    }
}
