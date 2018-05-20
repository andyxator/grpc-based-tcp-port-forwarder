#!/bin/sh
protoc -I/usr/local/include -I. -I$GOPATH/src --gofast_out=plugins=grpc:./ api.proto
