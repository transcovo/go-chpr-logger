#!/bin/bash

set -e


go get -t .

go vet $(go list ./... | grep -v /vendor/)

go get github.com/golang/lint/golint
golint -set_exit_status $(go list ./... | grep -v /vendor/)

for d in $(go list ./... | grep -v vendor); do
    go test $d
done
