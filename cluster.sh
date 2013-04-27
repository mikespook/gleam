#!/bin/sh
checkstatus()
{
    rc=1
    while [ $rc != 0 ] ; do
        echo 'Waiting...'
        sleep 1
        nc -z $1 $2 
        rc=$? 
    done
}
HOST=127.0.0.1
DZNS_PORT=9046
DZNS_WEB=9048
DZ_PORT=8047
DZ_WEB=8048

echo 'Starting DZNS...'
nohup doozerd -timeout 3 -l=$HOST:$DZNS_PORT -w=$HOST:$DZNS_WEB -c=dzns >/tmp/dzns.log 2>&1 &
checkstatus $HOST $DZNS_PORT
echo 'Done!'
echo 'Starting doozerd...'
nohup doozerd -timeout 3 -l=$HOST:$DZ_PORT -w=$HOST:$DZ_WEB -c=app -b=doozer:?ca=$HOST:$DZNS_PORT >/tmp/doozer.log 2>&1 &
checkstatus $HOST $DZ_PORT
echo 'Done!'

export CGO_CFLAGS="-I/usr/include/zookeeper"
export CGO_LDFLAGS="-lzookeeper_mt -lev"
