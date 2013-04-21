// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package main

import (
    "flag"
    "strings"
    "github.com/mikespook/golib/log"
)

var (
    logfile  = flag.String("log", "",
        "log to write " +
        "(empty for STDOUT)")
    loglevel = flag.String("log-level", "all", "log level " +
        "('error', 'warning', 'message', 'debug', 'all' and 'none'" +
        " are combined with '|')")
)

func init() {
    if flag.Parsed() {
        return
    }
    level := log.LogNone
    levels := strings.SplitN(*loglevel, "|", -1)
    for _, v := range levels {
        switch v {
        case "none":
            level = level | log.LogNone
            break
        case "error":
            level = level | log.LogError
        case "warning":
            level = level | log.LogWarning
        case "message":
            level = level | log.LogMessage
        case "debug":
            level = level | log.LogDebug
        case "all":
            level = log.LogAll
        default:
        }
    }
    if err := log.Init(*logfile, level); err != nil {
        log.Error(err)
    }
}
