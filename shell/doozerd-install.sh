#!/bin/bash

. common.sh

check_base_env

# 安装 Doozer 和依赖库
$GOBIN/go get github.com/ha/doozerd
