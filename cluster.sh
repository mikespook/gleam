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
echo 'Starting DZNS...'
nohup doozerd -timeout 5 -l ':10000' -w ':8000' -c 'dzns' >/tmp/dzns.log 2>&1 &
checkstatus 127.0.0.1 10000
echo 'Done!'
echo 'Starting doozerd...'
nohup doozerd -timeout 5 -l ':8046' -w ':8001' -c 'skynet' -b 'doozer:?ca=:10000' >/tmp/doozer.log 2>&1 &
checkstatus 127.0.0.1 8046
echo 'Done!'
