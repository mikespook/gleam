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
    dzuri = flag.String("doozer", "", "address of the doozerd, must be specified as `cn` when -dzns was assigned")
    dzburi = flag.String("dzns", "", "address of the DzNS")
    zk = flag.String("zk", "", "address of the ZooKeeper (using ',' as the separator for multi-ZooKeepers), one of -doozer and -zk must be specified")
    region = flag.String("region", node.DefaultRegion, "a region of the z-node located in (using ':' as the separator for multi-regions)")
    scriptPath = flag.String("script", "", "default script path(as the enviroment variable $Z_NODE_SCRIPT_ROOT)")
    encoding = flag.String("encoding", "json", "encoding of task message (JSON as default)")
)

func init() {
    if !flag.Parsed() {
        flag.Parse()
    }

    if *dzburi != "" && *dzuri == "" {
        flag.Usage()
        os.Exit(-1)
        return
    }

    if *dzuri == "" && *zk == "" {
        flag.Usage()
        os.Exit(-1)
        return
    }

    if *scriptPath == "" {
        *scriptPath = os.Getenv(SCRIPT_ROOT)
    }
    if *scriptPath == "" {
        var err error
        *scriptPath, err = os.Getwd()
        if err != nil {
            log.Error(err)
            os.Exit(-1)
            return
        }
    }
    node.SetDefaultPath(*scriptPath)

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
    hostname, err := os.Hostname()
    if err != nil {
        log.Error(err)
        return
    }
    n := node.New(hostname, strings.Split(*region, ":") ... )
    n.ErrHandler = node.ErrHandler

    if err := node.PyInit(); err != nil {
        log.Error(err)
        return
    }
    defer node.PyClose()
    n.ScriptHandler = node.PyExec

    if *dzuri != "" {
        log.Messagef("Connect to Doozerd: dzns=[%s], doozer=[%s]", *dzburi, *dzuri)
        d, err := node.NewDoozer(*dzuri, *dzburi)
        if err != nil {
            log.Error(err)
            return
        }
        n.AddConn(d)
    }

    if *zk != "" {
        log.Messagef("Connect to ZooKeeper: zk=[%s]", *zk)
        z, err := node.NewZooKeeper(*zk)
        if err != nil {
            log.Error(err)
            return
        }
        n.AddConn(z)
    }

    switch *encoding {
        case "gob":
            n.DecodeHandler = node.GobDecoder
        case "json":
            fallthrough
        default:
            n.DecodeHandler = node.JSONDecoder
    }

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
