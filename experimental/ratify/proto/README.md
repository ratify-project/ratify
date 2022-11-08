# Overview

| package | description |
| --- | --- |
| common | common protos that are imported by multiple packages |
| referrerstore | Referrer Store plugin service and associated protobuf messages |
| verifier | Verifier plugin service and associated protobuf messages |
| orchestrator | Orchestrator service and associated protobuf messages <br/> _Enables decoupling verifiers and referrer stores by having the orchestrator act as a passthrough_ |

These proto files enable development of microservice plugins. gRPC plugins are only supported when Ratify is running as a standalone server and **not** as an executable/binary.

When plugins are registered with Ratify, the configuration must provide the information required for Ratify to successfully connect to said service. Although an initial connection check will be performed when Ratify instantiates the client for a given plugin, the operator is responsible for configuring appropriate liveness and readiness probes. Any exceptions arising from connection issues shall bubble up and be included in Ratify error logs. 

At this time, there is no plans to include configurable retry logic nor shall a plugin be disabled after _x_ number of connection-related failures. 

# Code Generation

## Go example
1. Install the required packages
```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```
2. To Generate Code from Proto files  
```
cd ./experimental
./generate-protos.sh
```
3. Validate with 
```sh
go mod tidy
```
4. _**Depending on the value of "go_package" for each proto file, you may have to move the generated code to an alternate directory.**_ _e.g. setting the `go_package` to "github.com/deislabs/ratify/experimental/proto/v1/referrerstore" will create such a directory structure. Moving the generated code to {root}/experimental/proto/v1/referrerstore will resolve errors since the project's module is "github.com/deislabs/ratify"_
