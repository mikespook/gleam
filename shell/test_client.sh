#!/bin/bash

# etcd and gleam should be run manually

set -e

# call config::sync
pushd . > /dev/null
cd ../cmd/gleam-client/
go build

./gleam-client region
./gleam-client node
./gleam-client info
# call command see test_config_sync.sh

popd > /dev/null

exit 0
