#!/bin/bash

set -e

go get -t .

go vet $(go list ./... | grep -v /vendor/)

for d in $(go list ./... | grep -v vendor); do
    go test $d
done
