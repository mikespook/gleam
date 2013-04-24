package main

import (
    "flag"
    "time"
    "encoding/json"
    "github.com/mikespook/golib/log"
    "github.com/ha/doozer"
    "github.com/mikespook/z-node/node"
)

var (
    uri = flag.String("doozer", "doozer:?ca=127.0.0.1:8046", "address of the doozerd")
    buri = flag.String("dzns", "", "address of the DzNS")
    region = flag.String("region", "default", "a region of the z-node located in")
    host = flag.String("host", "localhost", "hostname of z-node")
    pid = flag.Int("pid", 0, "pid of z-node")
    fn = flag.String("func", "", "function name (must be specified)")
)

func init() {
    if !flag.Parsed() {
        flag.Parse()
    }
}

func main() {
    if *fn == "" {
        flag.Usage()
        return
    }
    conn, err := doozer.DialUri(*uri, *buri)
    if err != nil {
        log.Error(err)
        return
    }
    defer conn.Close()
    rev, err := conn.Rev()
    if err != nil {
        log.Error(err)
        return
    }
    var path string
    if *pid == 0 {
       path = node.MakeWire(*region)
    } else {
       path = node.MakeNode(node.NodeFile, *host, *pid)
    }
    params := make([]interface{}, flag.NArg())
    for i := 0; i < flag.NArg(); i ++ {
        params[i] = interface{}(flag.Arg(i))
    }
    f := &node.ZFunc{Name: *fn, Params: params}
    body, err := json.Marshal(f)
    if err != nil {
        log.Error(err)
        return
    }
    rev, err = conn.Set(path, rev, body)
    if err != nil {
        log.Error(err)
        return
    }
    log.Messagef("Rev: %d", rev)
    time.Sleep(time.Second)
}
