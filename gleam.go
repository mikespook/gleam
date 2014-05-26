package gleam

import (
	"fmt"
	"sync"

	"github.com/coreos/go-etcd/etcd"
	"github.com/mikespook/golib/iptpool"
)

const (
	Root       = "/gleam"
	RegionDir  = Root + "/region"
	RegionFile = RegionDir + "/%s"
	NodeDir    = Root + "/node"
	NodeFile   = NodeDir + "/%s"
	InfoDir    = Root + "/info"
	InfoFile   = InfoDir + "/%s"
	QUEUE_SIZE = 16
)

type Gleam struct {
	ErrHandler ErrorHandlerFunc
	iptPool    *iptpool.IptPool
	w          sync.WaitGroup

	closeChans map[string]chan bool
	fChan      chan *Func

	client *etcd.Client
}

func New(machines []string, script string, cert, key, ca string) (gleam *Gleam, err error) {
	gleam = &Gleam{
		closeChans: make(map[string]chan bool, 2),
		fChan:      make(chan *Func, 32),
		iptPool:    iptpool.NewIptPool(NewLuaIpt),
	}
	gleam.iptPool.OnCreate = func(ipt iptpool.ScriptIpt) error {
		ipt.Init(script)
		return nil
	}
	if cert != "" && key != "" && ca != "" {
		if gleam.client, err = etcd.NewTLSClient(machines, cert, key, ca); err != nil {
			return
		}
	} else {
		gleam.client = etcd.NewClient(machines)
	}
	return
}

func (gleam *Gleam) Serve() (err error) {
	for f := range gleam.fChan {
		go func(f *Func) {
			ipt := gleam.iptPool.Get()
			defer gleam.iptPool.Put(ipt)
			if err := ipt.Exec(f.Name, f.Data); err != nil {
				gleam.err(err)
				return
			}
		}(f)
	}
	return nil
}

func (gleam *Gleam) Watch(file string) {
	gleam.closeChans[file] = make(chan bool)
	go func() {
		rc := make(chan *etcd.Response)
		go func() {
			for r := range rc {
				if f, err := marshal(r.Node.Value); err != nil {
					gleam.err(err)
				} else {
					gleam.fChan <- f
				}
			}
		}()
		if _, err := gleam.client.Watch(file, 0, false, rc, gleam.closeChans[file]); err != nil {
			gleam.err(err)
		}
	}()
}

func (gleam *Gleam) Close() error {
	for _, c := range gleam.closeChans {
		close(c)
	}
	gleam.w.Wait()
	return nil
}

func (gleam *Gleam) err(err error) {
	if gleam.ErrHandler != nil {
		gleam.ErrHandler(err)
	}
}

func MakeRegion(region string) string {
	return fmt.Sprintf(RegionFile, region)
}

func MakeNode(id string) string {
	return fmt.Sprintf(NodeFile, id)
}
