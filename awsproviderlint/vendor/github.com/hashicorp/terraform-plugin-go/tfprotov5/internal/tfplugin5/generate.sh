#!/bin/bash

# We do not run protoc under go:generate because we want to ensure that all
# dependencies of go:generate are "go get"-able for general dev environment
# usability. To compile all protobuf files in this repository, run
# "make protobuf" at the top-level.

set -eu

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

cd "$DIR"

protoc --proto_path=. --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative tfplugin5.proto
