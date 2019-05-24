#!/bin/bash

# considering that you have git clone from repo and stored into /tmp folder

# add user to run service 'docnogen'
sudo useradd docnogen -s /sbin/nologin -M

# move service file
cd /tmp/go-kit-documentnogen/deployment/dev

sudo cp docnogen-api.service /etc/systemd/system/
sudo chmod 755 /etc/systemd/system/docnogen-api.service

echo "service moved"

# copy source code to go folder
mkdir -p /home/appadmin/go/src/github.com/howlun/go-kit-documentnogen
cd /tmp/go-kit-documentnogen
rm -rf /home/appadmin/go/src/github.com/howlun/go-kit-documentnogen/*
cp -r * /home/appadmin/go/src/github.com/howlun/go-kit-documentnogen/
cd /home/appadmin/go/src/github.com/howlun/go-kit-documentnogen/cmd/server
/usr/local/go/bin/go build

echo "building source..."

sleep 10s

echo "source built"

# enable and start the service
sudo systemctl daemon-reload

sudo systemctl enable docnogen-api
sudo systemctl restart docnogen-api

echo "deployment completed"