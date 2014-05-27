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

    $ go get github.com/mikespook/gleam/cmd/gleam
    $ $GOBIN/gleam

Witch takes the following flags:

 * -ca-file="": Path to the CA file
 * -cert-file="": Path to the cert file
 * -config="": Path to configuration file
 * -etcd="http://127.0.0.1:4001": A comma-delimited list of etcd
 * -key-file="": Path to the key file
 * -log="": log to write (empty for STDOUT)
 * -log-level="all": log level ('error', 'warning', 'message', 'debug', 'all' and 'none' are combined with '|')
 * -name="$HOST-$PID": Name of this node, `$HOST-$PID` will be used as default.
 * -pid="": PID file
 * -region="default": A comma-delimited list of regions to watch
 * -script="": Directory of lua scripts

_The configuration settings will cover flags._

Client
------

Gleam supplies a package

    $ go get github.com/mikespook/gleam

And a cli command

    $ go get github.com/mikespook/gleam/cmd/gleam-client
    $ $GOBIN/gleam-client

To operate Gleam nodes.

You can read [client's source code][client-src] for the package's usage.

The cli command takes the following flags:

 * -ca-file="": Path to the CA file
 * -cert-file="": Path to the cert file
 * -etcd="http://127.0.0.1:4001": A comma-delimited list of etcd
 * -key-file="": Path to the key file
 * -log="": log to write (empty for STDOUT)
 * -log-level="all": log level ('error', 'warning', 'message', 'debug', 'all' and 'none' are combined with '|')

And commands include :

 * call: Call a function on nodes file
 * region: List all regions
 * node: List all nodes
 * info: List all nodes info

See [shell/test\_*.sh][shell] for more information.

Case study
==========

Let's see a case for synchronizing configurations.

Assume we have a cluster witch need to synchronize thire crontab configurations.
In an old school way, we may use `rsync`, `scp` or something else to synchronize the configuration from one server to the others.
Through `gleam` we just need some steps to complate this job:

 1. `etcd` instances are running on systems in a same cluster;
 2. `gleam` nodes connected to the `etcd` cluster should be watching a same region(Eg. `default`);
 3. The configuration content has been writen to a file in `etcd`.
 4. Tell `gleam` call the lua script config::sync for synchronizing configuration.
 5. Done. 

See [test\_config\_sync.sh][config-sync] for more details.

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
 [config-sync]: https://github.com/mikespook/gleam/blob/master/shell/test_config_sync.sh
