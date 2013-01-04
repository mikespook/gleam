package node

import (
    "os"
    "fmt"
    "sync"
    "time"
    "encoding/json"
    "github.com/skynetservices/doozer"
    "github.com/mikespook/golib/funcmap"
)

const (
    MaxErrorCount = 5
    FuncPoolSize = 30
    WireFile = "%s/wire"
    NodeFile = "%s/node/%d"
    NodeList = "%s/info/%d"
    DefaultRegion = "/z-node"
)

type ZNode struct {
    ErrHandler ErrorHandlerFunc

    conn *doozer.Conn
    uri, buri, region string
    pid int
    rev int64
    fmap funcmap.Funcs
    w sync.WaitGroup
}

type ZFunc struct {
    Name string
    Params []interface{}
}

func New(region string) (node *ZNode) {
    if region == "" {
        region = DefaultRegion
    }
    if region[0] != '/' {
        region = "/" + region
    }
    return &ZNode {
        region: region,
        pid: os.Getpid(),
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
        fmt.Sprintf(NodeList, node.region, node.pid),
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
        fmt.Sprintf(NodeList, node.region, node.pid),
        node.rev); err != nil {
        node.err(err)
    }
    if err := node.conn.Del(
        fmt.Sprintf(NodeFile, node.region, node.pid),
        node.rev); err != nil {
        node.err(err)
    }
    node.conn.Close()
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
        if ev.IsDel() {
            break
        }
        if ev.IsSet() {
            var fn ZFunc
            if err := json.Unmarshal(ev.Body, &fn); err != nil {
                node.err(err)
                continue
            }
            if _, err := node.fmap.Call(fn.Name, fn.Params ...); err != nil {
                node.err(err)
                continue
            }
            i = 0
        }
    }
}

func (node *ZNode) watchSelf() {
    node.w.Add(1)
    go node.watch(fmt.Sprintf(NodeFile, node.region, node.pid))
}

func (node *ZNode) watchWire() {
    node.w.Add(1)
    go node.watch(fmt.Sprintf(WireFile, node.region))
}
