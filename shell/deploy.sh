#!/bin/bash

[ $EUID -ne 0 ] && echo 'root needed' && exit 1

cp init.d_z-node /etc/init.d/z-node
cp z-node.conf /etc/z-node.conf
mkdir -p /usr/share/z-node

update-rc.d -f z-node remove
update-rc.d z-node defaults
service z-node restart
