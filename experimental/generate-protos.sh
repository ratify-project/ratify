#! /bin/bash


## The protocol compiler searches for imported files in a set of directories specified on the protocol compiler command line using the -I/--proto_path flag. If no flag was given, it looks in the directory in which the compiler was invoked. In general you should set the --proto_path flag to the root of your project and use fully qualified names for all imports.


protoc \
--proto_path=./ratify/proto/v1 \
--go_out=. \
--go_opt=module=github.com/deislabs/ratify/experimental \
--go-grpc_out=. \
--go-grpc_opt=module=github.com/deislabs/ratify/experimental \
./ratify/proto/v1/*.proto
