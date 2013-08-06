#!/bin/bash

usage() {
    echo "Usage: setup.sh [-h host] [-p port] [-u user]"
    exit 1
}

while getopts 'h:p:u:' o &>> /dev/null; do
    case "$o" in
    h)
        HOST="$OPTARG";;
    p)
        PORT="$OPTARG";;
    u)
        U="$OPTARG";;
    *)
        usage;;
    esac
done

if [ "$HOST" == "" ]; then
    HOST=localhost
fi

if [ "$PORT" == "" ]; then
    PORT=22
fi

if [ "$U" == "" ]; then
    U=root
fi

ROOT_PATH=$(dirname $0)
# Load commons: functions, consts, variables
. lib/common.sh

# tar node dir
f=$(mktemp -u).tar.gz
tar zcvf $f * > /dev/null
scp -P $PORT $f $U@$HOST:$f && rm $f
d=/tmp/$(mktemp -u)
ssh -p $PORT $U@$HOST "rm -rf $d && mkdir -p $d && \
    tar zxvf $f -C $d >> /dev/null && rm $f && mv $d $HOME/"
