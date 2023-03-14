#!/usr/bin/env bash
##--------------------------------------------------------------------
#
# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
##--------------------------------------------------------------------

set -o errexit
set -o nounset
set -o pipefail

CERT_DIR=$1

generate() {
    # generate CA key and certificate
    echo "Generating CA key and certificate for ratify..."
    openssl genrsa -out ca.key 2048
    openssl req -new -x509 -days 1 -key ca.key -subj "/O=Ratify/CN=Ratify Root CA" -out ca.crt -addext "keyUsage=critical,keyCertSign"

    # generate leaf key and certificate
    echo "Generating leaf key and certificate for ratify..."
    openssl genrsa -out leaf.key 2048
    openssl req -newkey rsa:2048 -nodes -keyout leaf.key -subj "/CN=ratify.default" -out leaf.csr
    openssl x509 -req -extfile <(printf "basicConstraints=critical,CA:FALSE\nkeyUsage=critical,digitalSignature") -days 365 -in leaf.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out leaf.crt
}

rm -r ${CERT_DIR} || true
mkdir -p ${CERT_DIR}
pushd "${CERT_DIR}"
generate
popd
