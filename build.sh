#!/bin/bash
set -xeu
/usr/local/go/bin/go fmt rpiapi.go
env GOOS=linux GOARCH=arm GOARM=5 /usr/local/go/bin/go build rpiapi.go
