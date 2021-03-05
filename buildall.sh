#!/usr/bin/env bash

if [ ! -d out ]; then
    mkdir out
fi

TARGETS=$(go tool dist list)

for line in ${TARGETS}; do
    GOOS=$(echo ${line} | cut -d '/' -f 1)
    GOARCH=$(echo ${line} | cut -d '/' -f 2)

    export GOOS
    export GOARCH
    echo "Building adcscli-${GOOS}-${GOARCH}"
    go build -o out/adcscli-${GOOS}-${GOARCH} cli/adcscli/main.go
done

cd out
sha256sum *