#!/bin/bash

. common.sh

check_base_env

if [[ $EUID -e 0 ]]; then
    apt-get install -y gcc libc6-dev mercurial
    if [ "$?" != "0" ]; then 
    	echo "error, check your network!"
    	exit 1
    fi
fi

# init the environment
mkdir -p $SERVICE_BASE/bin
mkdir -p $SERVICE_BASE/go/own
mkdir -p $SERVICE_BASE/go/3rdpkg
# setup env-variables
export PATH=$PATH:$SERVICE_BASE/bin
export GOROOT=$SERVICE_BASE/go/golang
export GOBIN=$SERVICE_BASE/bin
# put $SERVICE_BASE/go/3rdpkg at the first
export GOPATH=$SERVICE_BASE/go/3rdpkg:$SERVICE_BASE/go/own:$GOROOT
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
echo 'export GOROOT=$SERVICE_BASE/go/golang' >> $config
echo 'export GOPATH=$SERVICE_BASE/go/3rdpkg:$SERVICE_BASE/go/own:$GOROOT' >> $config
echo 'export GOTOOLDIR=$GOROOT/pkg/tool' >> $config
echo 'export PATH=$PATH:$GOBIN' >> $config

cd $SERVICE_BASE/go
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
Golang was installed at $SERVICE_BASE/go/golang.
-----
done!
-----
EOF
exit 0
