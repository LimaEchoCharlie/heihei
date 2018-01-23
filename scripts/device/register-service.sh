#!/usr/bin/env bash

# script run on the device to register the service

SERVICE=$1 # service name

# give unit file the right permissions and move to correct directory
sudo chmod 644 $SERVICE
sudo mv $SERVICE /lib/systemd/system/$SERVICE

# tell systemd to start the service during the boot sequence
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE

# reboot system
sudo reboot
