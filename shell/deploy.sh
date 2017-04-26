#!/bin/bash

[ $EUID -ne 0 ] && echo 'root needed' && exit 1

cp gleam /usr/local/bin/
cp init.d_gleam /etc/init.d/gleam
mkdir -p /usr/local/share/gleam
cp -r scripts/* /usr/local/share/gleam/

update-rc.d -f gleam remove
update-rc.d gleam defaults
service gleam restart
