Z-Node
======

Z-Node is a task cluster, working with [Doozer cluster](https://github.com/ha/doozerd) or [ZooKeeper cluster](http://zookeeper.apache.org/). 

Every Z-Node watches two files:

 * /z-node/node/$HOST/$PID - for one-node tasks;

 * $REGION/wire - for cluster tasks.

Theory
======

Z-Node will register itself at `/z-node/info/$HOST/$PID` in Doozer. It watches the file `/z-node/node/$HOST/$PID`.
When the file was changed, Z-Node will be notified.

All of Z-Nodes are watching the file `$REGION/wire`. When this file was changed, all of Z-Node also will be notified.

The message (the file contents) is json encoded data with the function name and paramaters.
Z-Node will call the function with the paramaters.

    {
        Name: (string),
        Params: [(interface{}) ...]
    }

Dependencies
============

[Doozer](https://github.com/ha/doozer) [Golib](https://github.com/mikespook/golib) [py](https://github.com/qiniu/py) [gozk](https://github.com/petar/gozk)

Installing & Running
====================

All scripts were put in the directory [shell](https://github.com/mikespook/z-node/tree/master/shell).

Server node

> $ go get github.com/mikespook/z-node

Client

> $ go get github.com/mikespook/z-node/client

Authors
=======

 * Xing Xing <mikespook@gmail.com> [Blog](http://mikespook.com) [@Twitter](http://twitter.com/mikespook)

Open Source - MIT Software License
==================================
Copyright (c) 2012 Xing Xing

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
