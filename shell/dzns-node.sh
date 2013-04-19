#!/bin/bash

NAME=dzns
DESC=dzns

[ -f config.sh ] || exit 0
. config.sh


test -x $DAEMON || exit 0

set -e

. /lib/lsb/init-functions

start_dzns_node() {
    nohup $DAEMON -timeout $TIMEOUT -a $DZNS_MASTER -l $DZNS_NODE -w $DZNS_NODE_WEB -c 'dzns'>>$DZNS_NODE_LOG 2>&1 &
    set_dzns_node
}

set_dzns_node() {
    sleep 1
    n=`$DOOZER -a="doozer:?ca=$DZNS_MASTER" find /ctl/cal|tail -n 1`
    n=`expr ${n##*\/} + 1`
    printf "" | $DOOZER -a="doozer:?ca=$DZNS_MASTER" set /ctl/cal/$n 0
    arr=(`$DOOZER -a="doozer:?ca=$DZNS_MASTER" find /ctl/node|awk '{split($0, node, "/");if (node[4]!="")print node[4];}'|uniq`)
    for key in "${!arr[@]}"; do
        val=`$DOOZER -a="doozer:?ca=$DZNS_MASTER" get /ctl/ns/dzns/${arr[$key]}`
        if [ ! -n "$val" ]; then
            val=`$DOOZER -a="doozer:?ca=$DZNS_MASTER" get /ctl/node/${arr[$key]}/addr`
            printf $val | $DOOZER -a="doozer:?ca=$DZNS_MASTER" set /ctl/ns/dzns/${arr[$key]} 0
        fi
    done
}

stop_dzns_node() {
    PID=`ps -AF |grep $DAEMON|grep dzns|grep "\-a"|awk '{print $2;}'`
    if [ -n "$PID" ]; then
        kill $PID
    fi
}

case "$1" in
	start)
		echo -n "Starting $DESC: "
                start_dzns_node
		echo "$NAME."
		;;
	stop)
		echo -n "Stopping $DESC: "
                stop_dzns_node
		echo "$NAME."
		;;
	restart|force-reload)
		echo -n "Restarting $DESC: "
                stop_dzns_node
        	sleep 1
                start_dzns_node
		echo "$NAME."
		;;
	*)
		echo "Usage: $NAME {start|stop|restart|force-reload}" >&2
		exit 1
		;;
esac

exit 0
