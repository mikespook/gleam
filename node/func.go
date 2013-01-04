package node

import (
    "os"
    "syscall"
    "github.com/mikespook/golib/signal"
)

func Stop() {
    if err := signal.Send(os.Getpid(), os.Interrupt); err != nil {
        _err(err)
    }
}

func Restart() {
    var attr os.ProcAttr
    attr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
    attr.Sys = &syscall.SysProcAttr{}
    _, err := os.StartProcess(os.Args[0], os.Args, &attr)
    if err != nil {
        _err(err)
    }
    Stop()
}
