Gleam
=====

[![Build Status][travis-img]][travis]

Gleam is a cluster for helping system operations. It works with [etcd][etcd].

Every Gleam watches at least two files:

 * /gleam/node/$ID - for one-node tasks;
 * /gleam/region/$REGION - for cluster tasks (Every Gleam instance can 
 watch multi-regions).

Gleam will register itself as file `/gleam/info/$ID` with running 
informations. It watches the file `/gleam/node/$ID` for one-node 
tasks. If the file was changed, Gleam will be notified.

Gleam nodes watch the file `/gleam/region/$REGION` for cluster tasks.
When the file is changed, all watching Gleam will be notified.

The message (the file contents) is [JSON][json] encoding data with 
the function name and data. Gleam calls the function with data.

    {
        Name: (string),
        Data: (interface{}),
    }

Dependencies
============
 
 * [etcd][etcd]
 
 * [Golib][golib]

 * [lua][lua-for-go]

 * [luar][luar]

Installing & Running
====================

All useful scripts were put at the directory [shell][shell].

Server node
-----------

Compile Gleam, and run it:

    $ go get github.com/mikespook/gleam
    $ $GOBIN/gleam

Witch takes the following flags:

 * -etcd	- url of etcd;
 * -id		- node id, If it is not assigned, `$HOST-$PID` will be used as default;
 * -region	- regions to watch, multi-regions splite by `:`;
 * -script	- directory of lua scripts;
 * -pid 	- PID file;
 * -log 	- log file;
 * -log\_level - log level.

Client
------

Gleam supplies a package

    $ go get github.com/mikespook/gleam/client

And a cli command

    $ go get github.com/mikespook/gleam/cmd/client
    $ $GOBIN/client

To operate Gleam nodes.

You can read [client's source code][client-src] for the package's usage.

The cli command takes the following flags:

 * -etcd - url of etcd;
 * -target - node id or regines, multi-targets splite by `:`;
 * -fn - function name;
 * -data

Authors
=======

 * Xing Xing <mikespook@gmail.com> [Blog](http://mikespook.com) [@Twitter](http://twitter.com/mikespook)

Open Source - MIT Software License
==================================

See LICENSE.

 [etcd]: https://github.com/coreos/etcd
 [client-src]: https://github.com/mikespook/gleam/tree/master/cmd/client
 [luar]: https://github.com/stevedonovan/luar
 [travis-img]: https://travis-ci.org/mikespook/gleam.png?branch=master
 [travis]: https://travis-ci.org/mikespook/gleam
 [json]: http://www.json.org/
 [golib]: https://github.com/mikespook/golib
 [lua-for-go]: https://github.com/aarzilli/golua/lua
 [shell]: https://github.com/mikespook/z-node/tree/master/shell
