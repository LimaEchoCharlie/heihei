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
    echo "go test"
    go test -race
    echo "devel build"
    go build -race -ldflags "-X main.devel=yes"
else
    echo "release build"
    GOOS=linux GOARCH=arm GOARM=7 go build
fi
