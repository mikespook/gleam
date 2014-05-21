#!/bin/bash

[ $EUID -ne 0 ] && echo 'root needed' && exit 1

go get github.com/mikespook/gleam
cp $GOBIN/gleam /usr/bin/
cp init.d_gleam /etc/init.d/gleam
cp gleam.conf /etc/gleam.conf
mkdir -p /usr/share/gleam

update-rc.d -f gleam remove
update-rc.d gleam defaults
service gleam restart
