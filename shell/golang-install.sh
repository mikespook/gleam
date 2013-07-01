#!/bin/bash

. common.sh

check_base_env

if [ $EUID -eq 0 ]; then
    apt-get install -y gcc libc6-dev mercurial
    if [ "$?" != "0" ]; then 
    	echo "error, check your network!"
    	exit 1
    fi
fi

BIN=$SERVICE_BASE/bin
OWN=$SERVICE_BASE/golang/own
PKG=$SERVICE_BASE/golang/3rdpkg

# init the environment
mkdir -p $BIN
mkdir -p $OWN
mkdir -p $PKG
# setup env-variables
export PATH=$PATH:$BIN
export GOROOT=$SERVICE_BASE/golang/go
export GOBIN=$BIN
# put $SERVICE_BASE/go/3rdpkg at the first
export GOPATH=$PKG:$OWN:$GOROOT
export GOTOOLDIR=$GOROOT/pkg/tool

if [ $SERVICE_BASE == $HOME ]; then
    config=$SERVICE_BASE/.profile
else
    PROFILE=/etc/profile.d
    if [ -d $PROFILE ]; then
        config=$PROFILE/golang.sh
    else
        config=/etc/profile
    fi
fi

sed -i -e "/^$/d" $config
sed -i -e "s/^export SERVICE_BASE.*//g" $config
sed -i -e "s/^export PATH=\$PATH:\$GOBIN.*//g" $config
sed -i -e "s/^export GOROOT.*//g" $config
sed -i -e "s/^export GOBIN.*//g" $config
sed -i -e "s/^export GOPATH.*//g" $config
sed -i -e "s/^export GOTOOLDIR.*//g" $config

echo "" >>$config
echo "export SERVICE_BASE=$SERVICE_BASE" >> $config
echo 'export GOBIN=$SERVICE_BASE/bin' >> $config
echo 'export GOROOT=$SERVICE_BASE/golang/go' >> $config
echo 'export GOPATH=$SERVICE_BASE/golang/3rdpkg:$SERVICE_BASE/golang/own:$GOROOT' >> $config
echo 'export GOTOOLDIR=$GOROOT/pkg/tool' >> $config
echo 'export PATH=$PATH:$GOBIN' >> $config

cd $SERVICE_BASE/golang
rm -rf golang
hg clone -u release https://code.google.com/p/go
if [ "$?" != "0" ]; then 
	echo "error, check your network."
	exit 1
fi

cd $GOROOT/src/
./all.bash

if [ "$?" != "0" ]; then 
	echo "error, build golang faild."
	exit 1
fi

echo <<EOF 
Golang was installed at $SERVICE_BASE/golang/go.
-----
done!
-----
EOF
exit 0
