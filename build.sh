#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o dynamotd-linux-amd64
GOOS=linux GOARCH=arm64 go build -o dynamotd-linux-arm64
GOOS=darwin GOARCH=amd64 go build -o dynamotd-macos-amd64
GOOS=darwin GOARCH=arm64 go build -o dynamotd-macos-arm64
