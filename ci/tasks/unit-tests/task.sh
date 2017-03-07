#!/bin/bash
set -eu

go get github.com/onsi/ginkgo/ginkgo
go get github.com/tools/godep

export GOPATH=$PWD/go:$GOPATH

cd go/src/github.com/pivotal-cf/downtimer

godep restore
ginkgo -r -cover -race
