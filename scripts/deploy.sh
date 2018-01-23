#!/usr/bin/env bash

# deploy the executable to the given pi and register as a daemon 

HOST=$1 # e.g. user@server

# local variables
SERVICE="heihei.service"

# move to script location
cd "${0%/*}"

# stop service on pi
ssh $HOST sudo systemctl stop $SERVICE 

# move exec and systemd unit file to pi
scp ../heihei $HOST.local:
scp ../init/$SERVICE $HOST.local:

# run install script on the pi
ssh $HOST 'bash -s' < device/register-service.sh $SERVICE

cd ~
