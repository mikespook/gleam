#!/bin/bash

PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin
NAME=dzns
DESC=dzns

[ -f config.sh ] || exit 0
. config.sh


test -x $DAEMON || exit 0

set -e

. /lib/lsb/init-functions

start_dzns_master() {
    nohup $DAEMON -timeout $TIMEOUT -l $DZNS_MASTER -w $DZNS_MASTER_WEB -c 'dzns'>>$DZNS_MASTER_LOG 2>&1 &
}

stop_dzns_master() {
    PID=`ps -AF |grep $DAEMON|grep dzns|grep -v "\-a"|awk '{print $2;}'`
    if [ -n "$PID" ]; then
        kill $PID
    fi
}

case "$1" in
	start)
		echo -n "Starting $DESC: "
                start_dzns_master
		echo "$NAME."
		;;
	stop)
		echo -n "Stopping $DESC: "
                stop_dzns_master
                echo "$NAME."
		;;
	restart|force-reload)
		echo -n "Restarting $DESC: "
                stop_dzns_master
		sleep 1
                start_dzns_master
		echo "$NAME."
		;;
	*)
		echo "Usage: $NAME {start|stop|restart|force-reload}" >&2
		exit 1
		;;
esac

exit 0
