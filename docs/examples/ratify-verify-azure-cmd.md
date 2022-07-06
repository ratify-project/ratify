# Notary v2 Signature Verification With ACR Using Ratify

## Binaries

### Notation

```bash
curl -Lo notation.tar.gz https://github.com/shizhMSFT/notation/releases/download/v0.7.0-shizh.2/notation_0.7.0-shizh.2_linux_amd64.tar.gz

tar xvzf notation.tar.gz -C ~/bin notation
```

### ORAS

```bash
curl -LO https://github.com/oras-project/oras/releases/download/v0.2.1-alpha.1/oras_0.2.1-alpha.1_linux_amd64.tar.gz
mkdir oras
tar -xvf ./oras_0.2.1-alpha.1_linux_amd64.tar.gz -C ./oras/
cp ./oras/oras ~/bin/oras
```

### Ratify

```bash
# TODO update according to release and copy the plugin to ~/.ratify/plugins path
curl -Lo ratify.tar.gz https://github.com/deislabs/ratify/releases/download/v0.0.1/ratify_0.0.1_linux_amd64.tar.gz
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
az acr create -n $ACR_NAME -g $(ACR_NAME)-acr --sku Premium
az configure --default acr=wabbitnetworks
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
> You can repeat step 3-4 to create multiple signatures to the image.

### Discover & Verify using Ratify

- Create a Ratify config with ACR as the signature store and notary v2 as the signature verifier.

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
                    "~/.config/notation/certificate/wabbit-networks.io.crt"
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

oras push $REPO \
  --artifact-type 'sbom/example' \
  --subject $IMAGE \
  -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
  ./sbom.json:application/json

```

- Sign the SBoM
```bash
# Capture the digest, to sign it
SBOM_DIGEST=$(oras discover -o json \
                --artifact-type sbom/example \
                -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
                $IMAGE | jq -r ".references[0].digest")

notation sign $REPO@$SBOM_DIGEST
```

### Discover & Verify using Ratify

- Create a Ratify config with ACR as the store for SBoMs, Scan results and their corresponding signatures. Also, plugin the verifier for SBoM and scan results in the config.

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
        "artifactVerificationPolicies": {
            "application/vnd.cncf.notary.v2.signature": "any",
            "sbom/example": "all"
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [
            {
                "name":"notaryv2",
                "artifactTypes" : "application/vnd.cncf.notary.v2.signature",
                "verificationCerts": [
                    "~/.config/notation/certificate/wabbit-networks.io.crt"
                  ]
            },
            {
                "name":"sbom",
                "artifactTypes" : "sbom/example",
                "nestedReferences": "application/vnd.cncf.notary.v2.signature"
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