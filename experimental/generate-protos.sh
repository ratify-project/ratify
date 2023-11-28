# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#! /bin/bash


## The protocol compiler searches for imported files in a set of directories specified on the protocol compiler command line using the -I/--proto_path flag. If no flag was given, it looks in the directory in which the compiler was invoked. In general you should set the --proto_path flag to the root of your project and use fully qualified names for all imports.


protoc \
--proto_path=./ratify/proto/v1 \
--go_out=. \
--go_opt=module=github.com/deislabs/ratify/experimental \
--go-grpc_out=. \
--go-grpc_opt=module=github.com/deislabs/ratify/experimental \
./ratify/proto/v1/*.proto
