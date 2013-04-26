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
    "github.com/mikespook/golib/funcmap"
)

const (
    MaxErrorCount = 5
    FuncPoolSize = 30
    DefaultRegion = "default"
    Root = "/z-node"
    WireFile = Root + "/%s/wire"
    NodeFile = Root + "/node/%s/%d"
    InfoFile = Root + "/info/%s/%d"
    QUEUE_SIZE = 16
)

type ZFuncCall func(string, ... interface{}) error

type ZNode struct {
    ErrHandler ErrorHandlerFunc
    ScriptHandler ZFuncCall

    conns []Conn
    watcher chan []byte

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
        watcher: make(chan []byte, QUEUE_SIZE),
        conns: make([]Conn, 0),
    }
}

func (node *ZNode) Bind(name string, fn interface{}) (err error) {
    return node.fmap.Bind(name, fn)
}

func (node *ZNode) AddConn(conn Conn) (err error) {
    if err = conn.Register(node.infoFile,
        []byte(time.Now().String()));
        err != nil {
        node.err(err)
        return
    }
    node.conns = append(node.conns, conn)
    return
}

func (node *ZNode) Start() {
    node.watchSelf()
    node.watchWire()
    go node.loop()
}

func (node *ZNode) loop() {
    for {
        if data, ok := <-node.watcher; ok {
            var fn ZFunc
            if err := json.Unmarshal(data, &fn); err != nil {
                node.err(err)
                continue
            }
            go node.Call(fn.Name, fn.Params ...)
        } else {
            break
        }
    }
}

func (node *ZNode) Close() {
    for _, c := range node.conns {
        if err := c.Close(); err != nil {
            node.err(err)
        }
    }
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
    for _, c := range node.conns {
        node.w.Add(1)
        defer node.w.Done()
        for i := 0; i < MaxErrorCount; i ++ {
            if err := c.Watch(file, node.watcher); err != nil {
                if err == ErrConnection {
                    break
                }
                node.err(err)
                continue
            }
            i = 0
        }
    }
}

func (node *ZNode) Call(name string, params ... interface{}) {
    if _, ok := node.fmap[name]; ok {
        if _, err := node.fmap.Call(name, params ...); err != nil {
            node.err(err)
        }
        return
    }
    if node.ScriptHandler != nil {
        if err := node.ScriptHandler(name, params ...); err != nil {
            node.err(err)
       }
    }
}

func (node *ZNode) watchSelf() {
    go node.watch(node.nodeFile)
}

func (node *ZNode) watchWire() {
    for _, v := range node.wires {
        go node.watch(v)
    }
}
