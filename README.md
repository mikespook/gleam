Z-Node
======

Z-Node is a task executing node based on [Doozer cluster](https://github.com/skynetservices/doozerd). It watch two files:

 * $REGION/node/$HOST/$PID - for one-node tasks;

 * $REGION/wire - for broadcasting tasks.

Theory
======

Z-Node will register itself at `$REGION/info/$HOST/$PID` in Doozer.
It watches the file `$REGION/node/$HOST/$PID`. If it was changed, Z-Node will notice it.

All of Z-Node are watching `$REGION/wire`. If this file was changed, all of Z-Node will notice it.

The message is json encoded data with function name and paramaters. Z-Node will call the function with the paramaters.

    {
        Name: (string),
        Params: [(interface{}) ...]
    }

Dependencies
============

[Doozer](https://github.com/skynetservices/doozer)

[Golib](https://github.com/mikespook/golib)    

Install
=======

Server node

> $ go get github.com/mikespook/z-node/server

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
