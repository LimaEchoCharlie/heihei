#!/usr/bin/env bash

# build the code for ARM 7 (e.g. raspberry pi)
GOOS=linux GOARCH=arm GOARM=7 go build
