#!/usr/bin/env bash

go build ../main.go
protoc --plugin=protoc-gen-custom=./main --custom_out=dart_client service.proto
protoc --twirp_out=. --go_out=. ./service.proto
dartfmt -w dart_client/service.dart