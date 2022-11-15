#!/bin/bash
set -euo pipefail

# uncomment the line below to debug the script
# set -o xtrace 

go mod tidy

# copy the default dev configuration
mkdir -p ~/.ratify
cp .devcontainer/config.default.json ~/.ratify/config.json

# create the sample cert so that the default launch task works by default
mkdir -p ~/.ratify/ratify-certs
cat <<EOF >~/.ratify/ratify-certs/wabbit-networks.io.crt
-----BEGIN CERTIFICATE-----
MIIDNjCCAh6gAwIBAgIQH775NeR9QruFa7zk7yzgWDANBgkqhkiG9w0BAQsFADAdMRswGQYDVQQD
ExJ3YWJiaXQtbmV0d29ya3MuaW8wHhcNMjIxMDIwMjEyMzU0WhcNMjMxMDIwMjEzMzU0WjAdMRsw
GQYDVQQDExJ3YWJiaXQtbmV0d29ya3MuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQDGLZtDqEJGdou/m2sPOQLMAaPSGTCaJ4uWw3J56GlD/L0ahIvzaj0LnL1ovRYerUvBemAJtUqy
FJHobSEUq+eXfpOMs9hbxY8bJ90YYMNh2CZ6vTIDgSKw7JowFJNN6FnjPEPfwV2xAi/CU9r1ijMU
5m5MImnzbj4JiikdBjtQRPolQNWLubDqCw5hceC2pgcKlpWM//8SodAqu1lPBtlG5BhzPL8bQMAo
hW8Qe+Jo2J/AxO7Zp/mJFP4ipMhfAHDcQ1CWC0PRIO0qk9JhtSD+o0gQyTXQwFcWi5PgSnMe3rSE
kmI8WMe+oz7DYDENGz2Pgy7IqtTAphiDycQW6G3ZAgMBAAGjcjBwMA4GA1UdDwEB/wQEAwIFoDAJ
BgNVHRMEAjAAMBMGA1UdJQQMMAoGCCsGAQUFBwMDMB8GA1UdIwQYMBaAFGFzdC0AfxDJkxM1ZIVx
BxTrEGvKMB0GA1UdDgQWBBRhc3QtAH8QyZMTNWSFcQcU6xBryjANBgkqhkiG9w0BAQsFAAOCAQEA
mmXh2QAH03O5+ADia1AXAb97SIAEVSQj6mgYN9E9Os0/3aXDYVLKVShA1OiBidr6Iya9epFBbpzc
vJv15bGDg+k5KYAsRYSzfQpILADMYW4T/E3uaFBpbt0GYiwTlpRSKnEvyEwf9cRZOouE9CesIn2L
L0j/+zhBb8RpB7UbrcEL0PXnn/UbnQKkkekpwEOV/USrpT2T/AwE2GUiAzaLj8ps2qgm/4fPj/uh
w2W7SbZafkVYxirVIdra1Yq2V6BvnnCkbnBLdjAHbQ48+BgiJJTk3QZFofP88tRjGHPYN3TH2Ycv
Ibo/J0urezLK/M8/QeMkkhY1WClsAi+xLhVWUw==
-----END CERTIFICATE-----
EOF

# ensure plugins dir exists
mkdir -p ~/.ratify/plugins

# symlink .ratify for easy access during development
ln -snf ~/.ratify .ratify
