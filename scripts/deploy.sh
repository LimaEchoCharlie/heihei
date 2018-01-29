#!/usr/bin/env bash

# deploy the executable to the given pi and register as a daemon 

HOST=$1 # e.g. user@server

# local variables
SERVICE_NAME="heihei"

# move to script location
cd "${0%/*}"

# stop service on pi and remove logfile
ssh $HOST sudo systemctl stop $SERVICE_NAME.service 
ssh $HOST sudo rm $SERVICE_NAME.log

# move exec and systemd unit file to pi
scp ../heihei $HOST.local:
scp ../init/$SERVICE_NAME.service $HOST.local:

# run install script on the pi
ssh $HOST 'bash -s' < device/register-service.sh $SERVICE_NAME.service

cd ~
