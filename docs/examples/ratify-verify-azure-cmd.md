# Notary v2 Signature Verification With ACR Using Ratify

## Install Notation, ORAS, and Ratify

### Notation

Install Notation v0.11.0-alpha.4 with plugin support from [Notation GitHub Release](https://github.com/notaryproject/notation/releases/tag/v0.11.0-alpha.4).

```bash
# Download the Notation binary
curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v0.11.0-alpha.4/notation_0.11.0-alpha.4_linux_amd64.tar.gz
# Extract it from the binary
tar xvzf notation.tar.gz
# Copy the Notation CLI to your bin directory
mkdir -p ~/bin && cp ./notation ~/bin
```

### ORAS

Install ORAS 0.15.0 on a Linux machine. You can refer to the [ORAS installation guide](https://oras.land/cli/) for details.

```bash
# Download the ORAS binary
curl -LO https://github.com/oras-project/oras/releases/download/v0.15.0/oras_0.15.0_linux_amd64.tar.gz
# Create a folder to extract the ORAS binary
mkdir -p oras-install/
tar -zxf oras_0.15.0_*.tar.gz -C oras-install/
# Copy the Notation CLI to your bin directory
mv oras-install/oras /usr/local/bin/
```

### Ratify

Install Ratify v1.0.0-alpha.3 from [Ratify GitHub Release](https://github.com/deislabs/ratify/releases/tag/v1.0.0-alpha.3).

```bash
# Download the Ratify binary
curl -Lo ratify.tar.gz https://github.com/deislabs/ratify/releases/download/v1.0.0-alpha.3/ratify_1.0.0-alpha.3_Linux_amd64.tar.gz
# Extract it from the binary and copy it to the bin directory
tar xvzf ratify.tar.gz -C ~/bin ratify
```

## Presets

### Set up ACR and Auth information
```bash
export ACR_NAME=wabbitnetworks
export REGISTRY=$ACR_NAME.azurecr-test.io
export REPO=${REGISTRY}/net-monitor
export IMAGE=${REPO}:v1


# Create an ACR
# Premium to use tokens
az acr create -n $ACR_NAME -g $ACR_NAME-acr --sku Premium
az configure --default acr=$ACR_NAME
az acr update --anonymous-pull-enabled true

# Using ACR Auth with Tokens
export NOTATION_USERNAME='wabbitnetworks-token'
export NOTATION_PASSWORD=$(az acr token create -n $NOTATION_USERNAME \
                    -r wabbitnetworks \
                    --scope-map _repositories_admin \
                    --only-show-errors \
                    -o json | jq -r ".credentials.passwords[0].value")

docker login $REGISTRY -u $NOTATION_USERNAME -p $NOTATION_PASSWORD
oras login $REGISTRY -u $NOTATION_USERNAME -p $NOTATION_PASSWORD
```
## Demo 1 :  Discover & Verify Signatures using Ratify

### Sign the image using ```notation```

1. We will build, push, sign the image in ACR
```bash
# Build, push, sign the image in ACR
echo $IMAGE
```
2.  Build and push the image
```bash
# build the image
docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main

# push the image
docker push $IMAGE
```
3.  Generate a test certificate
```bash
# Generate a test certificate
notation cert generate-test --default "wabbit-networks.io"
```
4. Sign the image
```bash
notation sign $IMAGE
```
5.  List the signatures with notation
```bash
# List the signatures
notation list $IMAGE
```
> You can repeat step 4-5 to create multiple signatures to the image.

### Discover & Verify using Ratify

- Create a Ratify config with ORAS as the signature store and notary v2 as the signature verifier.

```bash
cat <<EOF > ~/.ratify/config.json 
{ 
    "store": { 
        "version": "1.0.0", 
        "plugins": [ 
            { 
                "name": "oras"
            }
        ]
    },
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.cncf.notary.v2.signature": "any"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [
            {
                "name":"notaryv2",
                "artifactTypes" : "application/vnd.cncf.notary.v2.signature",
                "verificationCerts": [
                    "~/.config/notation/localkeys/wabbit-networks.io.crt"
                  ]
            }
        ]
        
    }
}
EOF
```
- Discover the signatures

```bash
# Query for the signatures
export IMAGE_DIGEST_REF=$(docker image inspect $IMAGE | jq -r '.[0].RepoDigests[0]')
ratify discover -s $IMAGE_DIGEST_REF
``` 
- Verify all signatures for the image

```bash
# Verify signatures
ratify verify -s $IMAGE_DIGEST_REF
```

## Demo 2 : Discover & Verify SBoMs, scan results using Ratify

### Generate, Sign, Push SBoMs, Scan results

- Push an SBoM
 
```bash
# Create, Push
echo '{"version": "0.0.0.0", "artifact": "'${IMAGE}'", "contents": "good"}' > sbom.json

oras attach $IMAGE \
  --artifact-type sbom/example \
  -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
  sbom.json:application/json
```

- Sign the SBoM
```bash
# Capture the digest, to sign it
SBOM_DIGEST=$(oras discover -o json \
                --artifact-type sbom/example \
                -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
                $IMAGE | jq -r ".referrers[0].digest")

notation sign $REPO@$SBOM_DIGEST
```

### Discover & Verify using Ratify

- Create a Ratify config with ORAS as the store for SBoMs, Scan results and their corresponding signatures. Also, plugin the verifier for SBoM and scan results in the config.

```bash
cat <<EOF > ~/.ratify/config.json 
{ 
    "store": { 
        "version": "1.0.0", 
        "plugins": [ 
            { 
                "name": "oras"
            }
        ]
    },
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.cncf.notary.v2.signature": "any",
                "sbom/example": "all"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [
            {
                "name":"notaryv2",
                "artifactTypes" : "application/vnd.cncf.notary.v2.signature",
                "verificationCerts": [
                    "~/.config/notation/localkeys/wabbit-networks.io.crt"
                  ]
            },
            {
                "name":"sbom",
                "artifactTypes" : "sbom/example",
                "nestedReferences": "application/vnd.cncf.notary.v2.signature"
            }
        ]
    }
}
EOF
```

- Discover the signatures

```bash
# Discover full graph of supply chain content
ratify discover -s $IMAGE_DIGEST_REF
``` 
- Verify the full graph of supply chain content

```bash
# Verify full graph
ratify verify -s $IMAGE_DIGEST_REF
```