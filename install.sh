#!/bin/bash
export CGO_CFLAGS="-I/usr/include/zookeeper"
export CGO_LDFLAGS="-lzookeeper_mt"
go install
