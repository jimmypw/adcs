#!/usr/bin/env bash

if [ ! -d out ]; then
    mkdir out
fi

TARGETS=$(go tool dist list)
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