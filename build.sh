#!/bin/sh

cd alerterserver
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build
mv -f alerterserver ../programe

cd ../webapp
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build
mv -f  webapp ../programe




