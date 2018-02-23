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
    go test -race -cover
    echo "devel build"
    go build -race
else
    echo "release build"
    go generate -tags 'rapi'
    GOOS=linux GOARCH=arm GOARM=7 go build -tags 'rapi'
fi
