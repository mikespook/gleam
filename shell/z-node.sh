#!/bin/bash

NAME=z-node
DESC=z-node

[ -f config.sh ] || exit 0
. config.sh


test -x $Z_NODE || exit 0

set -e

. /lib/lsb/init-functions

start_z_node() {
    nohup $Z_NODE -doozer="doozer:?ca=$DOOZER_NODE" -region 'app'>>$Z_NODE_LOG 2>&1 &
}

stop_z_node() {
    PID=`ps -AF |grep $Z_NODE|awk '{print $2;}'`
    if [ -n "$PID" ]; then
        kill $PID
    fi
}

case "$1" in
	start)
		echo -n "Starting $DESC: "
                start_z_node
		echo "$NAME."
		;;
	stop)
		echo -n "Stopping $DESC: "
                stop_z_node
		echo "$NAME."
		;;
	restart|force-reload)
		echo -n "Restarting $DESC: "
                stop_z_node
        	sleep 1
                start_z_node
		echo "$NAME."
		;;
	*)
		echo "Usage: $NAME {start|stop|restart|force-reload}" >&2
		exit 1
		;;
esac

exit 0
