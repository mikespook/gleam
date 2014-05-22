package gleam

import (
	"fmt"
	"sync"

	"github.com/mikespook/golib/iptpool"
)

const (
	Root       = "/gleam"
	RegineFile = Root + "/regine/%s"
	NodeFile   = Root + "/node/%s"
	InfoFile   = Root + "/info/%s"
	QUEUE_SIZE = 16
)

type Gleam struct {
	ErrHandler ErrorHandlerFunc
	iptPool    *iptpool.IptPool
	w          sync.WaitGroup
}

type Func struct {
	Name string
	Data interface{}
}

func MakeRegion(region string) string {
	return fmt.Sprintf(RegineFile, region)
}

func MakeNode(id string) string {
	return fmt.Sprintf(NodeFile, id)
}

func New(id, script string) (gleam *Gleam, err error) {
	return &Gleam{}, nil
}

func (gleam *Gleam) Serve() {

}

func (gleam *Gleam) Watch(file string) error {
	return nil
}

func (gleam *Gleam) Close() error {
	gleam.w.Wait()
	return nil
}

func (gleam *Gleam) err(err error) {
	if gleam.ErrHandler != nil {
		gleam.ErrHandler(err)
	}
}
