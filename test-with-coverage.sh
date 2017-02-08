#!/bin/bash

set -e

go get -t .

go vet $(go list ./... | grep -v /vendor/)

echo "mode: set" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -v -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out | grep -v "mode: set" >> coverage.txt
        rm profile.out
    fi
done
