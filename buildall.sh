#!/usr/bin/env bash

ARCHS=('amd64', '386', 'arm')
OSS=('linux','darwin','windows','netbsd')

mkdir out

for GOARCH in $ARCHS; do
    for GOOS in $OSS; do
        go build -o out/adcscli-${GOOS}-${GOARCH} cli/adcscli
    done
done