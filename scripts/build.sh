#!/usr/bin/env bash

for var in "$@"; do
    case "${var}" in
    devel)
        isDevel=true;;
    esac
done


if [ "$isDevel" = true ]; then
    echo "go test"
    go test
    echo "devel build"
    # call go generate
    go generate
    go build
else
    echo "release build"
    go generate -tags 'rapi'
    GOOS=linux GOARCH=arm GOARM=7 go build -tags 'rapi'
fi
