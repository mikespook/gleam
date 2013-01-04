package main

import (
    "os"
    "flag"
    "time"
    "github.com/mikespook/golib/log"
    "github.com/mikespook/z-node/node"
    "github.com/mikespook/golib/signal"
)

var (
    uri = flag.String("doozer", "doozer:?ca=127.0.0.1:8046", "address of the doozerd")
    buri = flag.String("dzns", "", "address of the DzNS")
    region = flag.String("region", "z-node", "a region of the z-node located in")
)

func init() {
    if !flag.Parsed() {
        flag.Parse()
    }
    node.ErrHandler = func(err error) {
        log.Error(err)
    }
}

func main() {
    defer func() {
        log.Message("Exit.")
        time.Sleep(time.Second)
    }()
    log.Message("Starting...")

    n := node.New(*region)
    n.ErrHandler = node.ErrHandler
    n.Bind("Stop", node.Stop)
    n.Bind("Restart", node.Restart)
//    n.Bind("Shell", ExecShell)
//    n.Bind("PHP", ExecPHP)
    if err := n.Start(*uri, *buri); err != nil {
        log.Error(err)
    }
    defer n.Close()
    go func() {
        n.Wait()
        signal.Send(os.Getpid(), os.Interrupt)
    }()
    // signal handler
    sh := signal.NewHandler()
    sh.Bind(os.Interrupt, func() bool {return true})
    sh.Loop()
}
