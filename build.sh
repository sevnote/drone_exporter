#!/bin/bash
bin_file="./bin/drone_exporter"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static" -s -w' -o ${bin_file} src/main.go || exit 1

#upx --brute  ${bin_file}
#kill -9 $(pgrep ${bin_file})
#nohup ./${bin_file} &
#echo $(pgrep ${bin_file})

