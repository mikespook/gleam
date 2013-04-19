#!/bin/bash

NAME=doozer
DESC=doozer

[ -f config.sh ] || exit 0
. config.sh


test -x $DAEMON || exit 0

set -e

. /lib/lsb/init-functions

start_doozer_node() {
    DZNS=$DZNS_MASTER
    if [ $DZNS != $DZNS_NODE ]; then
        DZNS=$DZNS\&$DZNS_NODE
    fi
    nohup $DAEMON -timeout $TIMEOUT -l $DOOZER_NODE -b="doozer:?ca=$DZNS" -w $DOOZER_NODE_WEB -c 'app'>>$DOOZER_NODE_LOG 2>&1 &
    set_doozer_node
}

set_doozer_node() {
    sleep 1
    n=`$DOOZER -b="doozer:?ca=$DZNS_MASTER" find /ctl/cal|tail -n 1`
    n=`expr ${n##*\/} + 1`
    printf "" | $DOOZER -b="doozer:?ca=$DZNS_MASTER" set /ctl/cal/$n 0
}


stop_doozer_node() {
    PID=`ps -AF |grep $DAEMON|grep app|awk '{print $2;}'`
    if [ -n "$PID" ]; then
        kill $PID
    fi
}

case "$1" in
	start)
		echo -n "Starting $DESC: "
                start_doozer_node
		echo "$NAME."
		;;
	stop)
		echo -n "Stopping $DESC: "
                stop_doozer_node
		echo "$NAME."
		;;
	restart|force-reload)
		echo -n "Restarting $DESC: "
                stop_doozer_node
        	sleep 1
                start_doozer_node
		echo "$NAME."
		;;
	*)
		echo "Usage: $NAME {start|stop|restart|force-reload}" >&2
		exit 1
		;;
esac

exit 0
