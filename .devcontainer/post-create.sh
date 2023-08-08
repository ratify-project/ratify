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
