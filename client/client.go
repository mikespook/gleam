package main

import (
    "fmt"
    "flag"
    "encoding/json"
    "github.com/mikespook/golib/log"
    "github.com/skynetservices/doozer"
    "github.com/mikespook/z-node/node"
)

var (
    uri = flag.String("doozer", "doozer:?ca=127.0.0.1:8046", "address of the doozerd")
    buri = flag.String("dzns", "", "address of the DzNS")
    region = flag.String("region", "z-node", "a region of the z-node located in")
    pid = flag.Int("pid", 0, "pid of z-node")
    fn = flag.String("func", "Stop", "function name")
)

func init() {
    if !flag.Parsed() {
        flag.Parse()
    }
}

func main() {
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
    if (*region)[0] != '/' {
        *region = "/" + *region
    }
    var path string
    if *pid == 0 {
       path = fmt.Sprintf("%s/wire", *region)
    } else {
       path = fmt.Sprintf("%s/node/%d", *region, *pid)
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
}
