#!/bin/bash

for os in linux darwin ; do
  GOOS=$os GOARCH=amd64 go build -o dynamotd-$os-amd64
done
