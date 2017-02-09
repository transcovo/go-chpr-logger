#!/bin/bash

set -e

go get github.com/robertkrimen/godocdown/godocdown

cat > README.md <<- END
[![CircleCI](https://circleci.com/gh/transcovo/go-chpr-logger.svg?style=shield)](https://circleci.com/gh/transcovo/go-chpr-logger)
[![codecov](https://codecov.io/gh/transcovo/go-chpr-logger/branch/master/graph/badge.svg)](https://codecov.io/gh/transcovo/go-chpr-logger)
[![GoDoc](https://godoc.org/github.com/transcovo/go-chpr-logger?status.svg)](https://godoc.org/github.com/transcovo/go-chpr-logger)

Doc below generated from godoc with godocdown (see dev-tools/test.sh)

--------------------
END

godocdown . >> README.md
