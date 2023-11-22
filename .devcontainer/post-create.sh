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

#!/bin/bash
set -euo pipefail

# uncomment the line below to debug the script
# set -o xtrace 

go mod tidy

# copy the default dev configuration if one does not exist
mkdir -p ~/.ratify
if [ ! -f ~/.ratify/config.json ]; then
  cp .devcontainer/config.default.json ~/.ratify/config.json
fi

# create the sample cert so that the default launch task works by default
mkdir -p ~/.ratify/ratify-certs/notation/truststore/x509/ca/certs
cp ./test/bats/tests/certificates/wabbit-networks.io.crt ~/.ratify/ratify-certs/notation/truststore/x509/ca/certs/wabbit-networks.io.crt

mkdir -p ~/.ratify/ratify-certs/cosign
cp ./test/bats/tests/certificates/cosign.pub ~/.ratify/ratify-certs/cosign/cosign.pub

# ensure plugins dir exists
mkdir -p ~/.ratify/plugins

# symlink .ratify for easy access during development
ln -snf ~/.ratify .ratify
