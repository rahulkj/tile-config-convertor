#!/bin/bash -e

set +e
go mod init
set -e

go mod verify
go mod tidy

go test

if [[ -d releases ]]; then
  rm -rf releases
fi

mkdir releases

GOOS=linux GOARCH=amd64 go build -o releases/tile-config-convertor_linux_amd64 github.com/rahulkj/tile-config-convertor

GOOS=darwin GOARCH=amd64 go build -o releases/tile-config-convertor-darwin_amd64 github.com/rahulkj/tile-config-convertor

GOOS=windows GOARCH=386 go build -o releases/tile-config-convertor-windows_amd64.exe github.com/rahulkj/tile-config-convertor
