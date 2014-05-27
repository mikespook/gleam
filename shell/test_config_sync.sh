#!/bin/bash

# etcd and gleam should be run manually

set -e

# target file for testing
target=`mktemp -u`
# source config file
source=/etc/crontab

# install etcdctl
go get -u github.com/coreos/etcdctl

# write configuration to etcd
$GOBIN/etcdctl set $target < $source > /dev/null

# call config::sync
pushd . > /dev/null
cd ../cmd/gleam-client/
go build
./gleam-client call /gleam/region/default config::sync $target
popd > /dev/null

# sync needs a little time to complate
i=0
while [ ! -f $target ]; do
	sleep 0.1
	# it's too long time
	if [ $i -gt 50 ]; then
		echo "[$target] sync is failure"
		exit -1		
	fi
	i=$(($i + 1))
done

echo "[$target] sync is success"
# should be no difference
diff $target $source
exit 0
