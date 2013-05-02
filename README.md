Z-Node
======

Z-Node is a cluster for helping operations. It works with 
[Doozer cluster](https://github.com/ha/doozerd) and 
[ZooKeeper cluster](http://zookeeper.apache.org/). 

Every Z-Node watches two files:

 * /z-node/node/$HOST/$PID - for one-node tasks;

 * /z-node/$REGION/wire - for cluster tasks.

Z-Node will register itself as file `/z-node/info/$HOST/$PID` with running 
informations. It watches the file `/z-node/node/$HOST/$PID` for one-node 
tasks. When the file was changed, Z-Node will be notified.

All of Z-Nodes watch the file `/z-node/$REGION/wire` for cluster tasks.
When the file is changed, all of Z-Nodes will be notified.

The message (the file contents) is encoded data (Default by 
[JSON](http://www.json.org/)) with the function name and paramaters. Z-Node 
will call the function with the paramaters.

    {
        Name: (string),
        Params: [(interface{}) ...]
    }

Dependencies
============

 * [Doozer](https://github.com/ha/doozer) 

 * [Golib](https://github.com/mikespook/golib)
 
 * [py](https://github.com/qiniu/py)

 * [gozk](https://github.com/petar/gozk)

Installing & Running
====================

All useful scripts were put at the directory [shell](https://github.com/mikespook/z-node/tree/master/shell).

Befor building, the proper ZooKeeper libraries and headers must be installed .
E.g. Ubuntu 12.10, the package `libzookeeper-mt-dev` must be installed. 

The ZooKeeper package use cgo to communicat with ZooKeeper server.

This two environment variables must be set (On Ubuntu 12.10):

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
    $ ./client -dzns="doozer:?ca=127.0.0.1:9046" -doozer="doozer:?cn=app" -func=test abc def 123 456
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
