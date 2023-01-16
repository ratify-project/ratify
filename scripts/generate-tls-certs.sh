#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ns=${2:-default}
CERT_DIR=$1

generate() {
    # generate CA key and certificate
    echo "Generating CA key and certificate for ratify..."
    openssl genrsa -out ca.key 2048
    openssl req -new -x509 -days 1 -key ca.key -subj "/O=Ratify/CN=Ratify Root CA" -out ca.crt

    # generate server key and certificate
    echo "Generating server key and certificate for ratify..."
    openssl genrsa -out server.key 2048
    openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/CN=ratify.${ns}" -out server.csr
    openssl x509 -req -extfile <(printf "subjectAltName=DNS:ratify.${ns}") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
}

rm -r ${CERT_DIR} || true
mkdir -p ${CERT_DIR}
pushd "${CERT_DIR}"
generate
popd
