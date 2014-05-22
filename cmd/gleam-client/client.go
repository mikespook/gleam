package main

import (
	"flag"
	"fmt"
	"github.com/ha/doozer"
	"github.com/mikespook/golib/log"
	"github.com/mikespook/z-node/node"
	zookeeper "github.com/petar/gozk"
	"os"
	"strings"
	"time"
)

var (
	dzuri    = flag.String("doozer", "", "address of the doozerd, must be specified as `cn` when -dzns was assigned")
	dzburi   = flag.String("dzns", "", "address of the DzNS")
	zkuri    = flag.String("zk", "", "address of the ZooKeeper (using ',' as the separator for multi-ZooKeepers), one of -doozer and -zk must be specified")
	region   = flag.String("region", "default", "a region of the z-node located in")
	host     = flag.String("host", "localhost", "hostname of z-node")
	pid      = flag.Int("pid", 0, "pid of z-node")
	fn       = flag.String("func", "", "function name (must be specified)")
	encoding = flag.String("encoding", "json", "encoding of task message (JSON as default)")
	coder    node.Encoding
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

	if *dzuri == "" && *zkuri == "" {
		flag.Usage()
		os.Exit(-1)
		return
	}

	if *fn == "" {
		flag.Usage()
		os.Exit(-1)
		return
	}
}

func main() {
	switch *encoding {
	case "gob":
		var c node.Gob
		coder = c
	case "json":
		fallthrough
	default:
		var c node.JSON
		coder = c
	}

	var path string
	if *pid == 0 {
		path = node.MakeWire(*region)
	} else {
		path = node.MakeNode(node.NodeFile, *host, *pid)
	}
	params := make(map[string]interface{}, flag.NArg())
	log.Debug(flag.Args())
	for i := 0; i < flag.NArg(); i++ {
		var key, value string
		arg := flag.Arg(i)
		if strings.Contains(arg, "=") {
			str := strings.SplitN(arg, "=", 2)
			key = str[0]
			value = str[1]
		} else {
			key = fmt.Sprintf("%d", i)
			value = arg
		}
		params[key] = interface{}(value)
	}

	if *dzuri != "" {
		doozerd(path, params)
	}

	if *zkuri != "" {
		zk(path, params)
	}
	time.Sleep(time.Second)
}

func doozerd(path string, params interface{}) {
	conn, err := doozer.DialUri(*dzuri, *dzburi)
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
	f := &node.ZFunc{Name: *fn, Params: params}
	body, err := coder.Encode(f)
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
}

func zk(path string, params interface{}) {
	conn, zch, err := zookeeper.Dial(*zkuri, 5e9)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()
	event := <-zch
	if event.State != zookeeper.STATE_CONNECTED {
		log.Errorf("Event state error: %d", event.State)
		return
	}

	f := &node.ZFunc{Name: *fn, Params: params}
	body, err := coder.Encode(f)
	if err != nil {
		log.Error(err)
		return
	}
	stat, err := conn.Exists(path)
	if err != nil {
		return
	}
	if stat, err = conn.Set(path, string(body), stat.Version()); err != nil {
		log.Error(err)
		return
	}
	log.Messagef("Rev: %d", stat.Version())
}
