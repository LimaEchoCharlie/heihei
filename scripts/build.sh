#!/usr/bin/env bash

for var in "$@"; do
    case "${var}" in
    devel)
        isDevel=true;;
    esac
done

# call go generate
go generate

if [ "$isDevel" = true ]; then
    echo "devel build"
    go build -ldflags "-X main.devel=yes"
else
    echo "release build"
    GOOS=linux GOARCH=arm GOARM=7 go build
fi
