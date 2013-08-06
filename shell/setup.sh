#!/bin/bash

[ $EUID -ne 0 ] || echo 'root needed' && exit 1

cp -R ./etc/ /
cp ./bin/* /usr/bin/
for $s in ./etc/init.d/*; do
    s=`basename $s`
    update-rc.d $s defaults
    service $s start
done
