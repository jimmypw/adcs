#!/usr/bin/env bash

if [ ! -d out ]; then
    mkdir out
fi

TARGETS=("darwin/amd64"
"darwin/arm64"
"freebsd/386"
"freebsd/amd64"
"freebsd/arm"
"freebsd/arm64"
"linux/386"
"linux/amd64"
"linux/arm"
"linux/arm64"
"linux/loong64"
"netbsd/386"
"netbsd/amd64"
"netbsd/arm"
"netbsd/arm64"
"openbsd/386"
"openbsd/amd64"
"openbsd/arm"
"openbsd/arm64"
"openbsd/mips64"
"windows/386"
"windows/amd64"
"windows/arm"
"windows/arm64")

VERSION=$(go run cli/adcscli/main.go -v)

for line in ${TARGETS}; do
    GOOS=$(echo ${line} | cut -d '/' -f 1)
    GOARCH=$(echo ${line} | cut -d '/' -f 2)
    if [ ${GOOS} == "windows" ]; then
        suffix='.exe'
    else
        suffix=''
    fi
    
    export GOOS
    export GOARCH
    echo "Building adcscli-${VERSION}-${GOOS}-${GOARCH}"
    go build -o "out/adcscli-${VERSION}-${GOOS}-${GOARCH}${suffix}" cli/adcscli/main.go
done

cd out
sha256sum *
