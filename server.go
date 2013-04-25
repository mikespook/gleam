package main

import (
    "os"
    "flag"
    "time"
    "strings"
    "github.com/mikespook/golib/log"
    "github.com/mikespook/z-node/node"
    "github.com/mikespook/golib/signal"
)

const (
    SCRIPT_ROOT = "Z_NODE_SCRIPT_ROOT"
)

var (
    dzuri = flag.String("doozer", "", "address of the doozerd")
    dzburi = flag.String("dzns", "", "address of the DzNS")
    zk = flag.String("zk", "", "address of the ZooKeeper")
    region = flag.String("region", node.DefaultRegion, "a region of the z-node located in (using ':' as the separator for multi-regions)")
    scriptPath = flag.String("script", "", "default script path(as the enviroment variable $Z_NODE_SCRIPT_ROOT)")
)

func init() {
    if !flag.Parsed() {
        flag.Parse()
    }
    node.ErrHandler = func(err error) {
        log.Error(err)
    }

    if *scriptPath == "" {
        *scriptPath = os.Getenv(SCRIPT_ROOT)
    }
    if *scriptPath == "" {
        var err error
        *scriptPath, err = os.Getwd()
        if err != nil {
            log.Error(err)
        }
    }
    node.SetDefaultPath(*scriptPath)
}

func main() {
    defer func() {
        log.Message("Exit.")
        time.Sleep(time.Second)
    }()
    log.Message("Starting...")
    hostname, err := os.Hostname()
    if err != nil {
        log.Error(err)
        return
    }
    n := node.New(hostname, strings.Split(*region, ":") ... )
    n.ErrHandler = node.ErrHandler
    d, err := node.NewDoozer(*uri, *buri)
    if err != nil {
        log.Error(err)
        return
    }
    n.AddConn(d)
    n.Bind("Stop", node.Stop)
    n.Bind("Restart", node.Restart)
    n.Start()
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
