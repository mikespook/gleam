package gleam

import (
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/mikespook/golib/iptpool"
)

const (
	Root        = "/gleam"
	RegionDir   = Root + "/region"
	RegionFile  = RegionDir + "/%s"
	NodeDir     = Root + "/node"
	NodeFile    = NodeDir + "/%s"
	InfoDir     = Root + "/info"
	InfoFile    = InfoDir + "/%s"
	InfoCreated = InfoFile + "/created"
	InfoRemoved = InfoFile + "/removed"
	InfoError   = InfoFile + "/error"
	QUEUE_SIZE  = 16
)

type Gleam struct {
	ErrHandler ErrorHandlerFunc
	iptPool    *iptpool.IptPool
	w          sync.WaitGroup

	id string

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
		ipt.Bind("Get", gleam.luaGet)
		ipt.Bind("GetDir", gleam.luaGetDir)
		ipt.Bind("Set", gleam.luaSet)
		ipt.Bind("Delete", gleam.luaDelete)
		ipt.Bind("Watch", gleam.luaWatch)
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
	if err = gleam.register(); err != nil {
		return
	}
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
	if err = gleam.unregister(); err != nil {
		return
	}
	return
}

func (gleam *Gleam) register() error {
	_, err := gleam.client.Set(fmt.Sprintf(InfoCreated, gleam.id), time.Now().String(), 0)
	return err
}

func (gleam *Gleam) unregister() error {
	if _, err := gleam.client.Set(fmt.Sprintf(InfoRemoved, gleam.id), time.Now().String(), 0); err != nil {
		return err
	}
	if _, err := gleam.client.SetDir(fmt.Sprintf(InfoFile, gleam.id), 5); err != nil {
		return err
	}
	if _, err := gleam.client.Delete(fmt.Sprintf(NodeFile, gleam.id), true); err != nil {
		return err
	}
	return nil
}

func (gleam *Gleam) WatchNode(id string) {
	gleam.id = id
	gleam.Watch(MakeNode(id))
}

func (gleam *Gleam) WatchRegion(region string) {
	gleam.Watch(MakeRegion(region))
}

func (gleam *Gleam) Watch(file string) {
	gleam.closeChans[file] = make(chan bool)
	gleam.w.Add(1)
	go func() {
		defer gleam.w.Done()
		rc := make(chan *etcd.Response)
		go func() {
			for r := range rc {
				if f, err := MarshalFunc(r.Node.Value); err != nil {
					gleam.err(err)
				} else {
					gleam.fChan <- f
				}
			}
		}()
		if _, err := gleam.client.Watch(file, 0, false, rc, gleam.closeChans[file]); err != nil {
			if err != etcd.ErrWatchStoppedByUser {
				gleam.err(err)
			}
		}
	}()
}

func (gleam *Gleam) Wait() {
	gleam.w.Wait()
}

func (gleam *Gleam) Close() error {
	for _, c := range gleam.closeChans {
		close(c)
	}
	close(gleam.fChan)
	gleam.w.Wait()
	return nil
}

func (gleam *Gleam) err(err error) {
	if gleam.ErrHandler != nil {
		gleam.ErrHandler(err)
	}
	if e, ok := err.(*etcd.EtcdError); ok && e.ErrorCode == etcd.ErrCodeEtcdNotReachable {
		return
	}
	gleam.client.Set(fmt.Sprintf(InfoError, gleam.id), err.Error(), 0)
}

func MakeRegion(region string) string {
	return fmt.Sprintf(RegionFile, region)
}

func MakeNode(id string) string {
	return fmt.Sprintf(NodeFile, id)
}
