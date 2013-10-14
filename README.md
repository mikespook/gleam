Z-Node
======

[![Build Status][travis-img]][travis]

Z-Node is a cluster for helping system operations. It works with 
[Doozer cluster][doozerd] and 
[ZooKeeper cluster][zk]. 

Every Z-Node watches at least two files:

 * /z-node/node/$HOST/$PID - for one-node tasks;

 * /z-node/$REGION/wire - for cluster tasks (Every Z-Node instance can 
 watch multi-regions).

Z-Node will register itself as file `/z-node/info/$HOST/$PID` with running 
informations. It watches the file `/z-node/node/$HOST/$PID` for one-node 
tasks. When the file was changed, Z-Node will be notified.

All of Z-Nodes watch the file `/z-node/$REGION/wire` for cluster tasks.
When the file is changed, all of Z-Nodes will be notified.

The message (the file contents) is encoded data (Default by 
[JSON][json]) with the function name and paramaters. Z-Node 
will call the function with the paramaters.

    {
        Name: (string),
        Params: [(interface{}) ...]
    }

Dependencies
============

 * [Doozer][doozer]

 * [Golib][golib]

 * [lua][lua-for-go]

 * [luar][luar]
 
 * [gozk][gozk]

 * libzookeeper-mt-dev and liblua5.1-0-dev for Ubuntu


Installing & Running
====================

All useful scripts were put at the directory [shell][shell].

Befor building, the proper ZooKeeper and lua libraries and headers must be installed .
E.g. Ubuntu 13.04, the package `libzookeeper-mt-dev` and `liblua5.1-0-dev` must be installed. 

The ZooKeeper package use cgo to communicat with ZooKeeper server.

This two environment variables must be set (tested on Ubuntu):

    $ export CGO_CFLAGS="-I/usr/include/zookeeper"
    $ export CGO_LDFLAGS="-lzookeeper_mt"

Server node
-----------

    $ cd github.com/mikespook/z-node
    $ go build
    $ ./z-node -dzns="doozer:?ca=127.0.0.1:9046" -doozer="doozer:?cn=app" -script="./script"
    $ ./z-node -zk="127.0.0.1:2181" -script="./script" -region="op:testing:backup"

__Note__

 1. When `-dzns` was assigned, `-doozer` also must be specified.

 2. Use Doozer's URI format for doozer's connection.
    * `doozer:?ca=$IP:$PORT`: both for `-dzns` and `-doozer`
    * `doozer:?cn=$CLUSTER_NAME`: `-doozer` only

 3. Use ':' as the separator for multi-regions. E.g. `-region=a:b:c` specified 3 regions "a", "b" and "c".

 4. Set the enviroment variable `$Z_NODE_SCRIPT_ROOT` for the Z-Node's script searching path. The param `-script` will recover this variable. If both of them were empty, the current directory was set as default.

 5. Params `-doozer` and `-zk` must be specified one or either.

Client
------

    $ cd github.com/mikespook/z-node/client
    $ go build
    $ ./client -dzns="doozer:?ca=127.0.0.1:9046" -doozer="doozer:?cn=app" -func=test abc def foobar=456
    $ ./client -zk="127.0.0.1:2181" -func=test -region=testing

Authors
=======

 * Xing Xing <mikespook@gmail.com> [Blog](http://mikespook.com) [@Twitter](http://twitter.com/mikespook)

Open Source - MIT Software License
==================================
Copyright (c) 2012 Xing Xing

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

 [luar]: https://github.com/stevedonovan/luar
 [doozerd]: https://github.com/ha/doozerd
 [doozer]: https://github.com/ha/doozer
 [zk]: http://zookeeper.apache.org
 [travis-img]: https://travis-ci.org/mikespook/z-node.png?branch=master
 [travis]: https://travis-ci.org/mikespook/z-node
 [json]: http://www.json.org/
 [golib]: https://github.com/mikespook/golib
 [lua-for-go]: https://github.com/aarzilli/golua/lua
 [gozk]: https://github.com/petar/gozk
 [shell]: https://github.com/mikespook/z-node/tree/master/shell
