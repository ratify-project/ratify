# Notary v2 Signature Verification With ACR Using Ratify

## Install Notation, ORAS, and Ratify

### Notation

Install Notation v1.0.0-rc.1 with plugin support from [Notation GitHub Release](https://github.com/notaryproject/notation/releases/tag/v1.0.0-rc.1).

```bash
# Download the Notation binary
curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v1.0.0-rc.1/notation_1.0.0-rc.1_linux_amd64.tar.gz
# Extract it from the binary and copy it to the bin directory
tar xvzf notation.tar.gz -C  /usr/local/bin notation
```

### ORAS

Install ORAS 0.16.0 on a Linux machine. You can refer to the [ORAS installation guide](https://oras.land/cli/) for details.

```bash
# Download the ORAS binary
curl -Lo oras.tar.gz https://github.com/oras-project/oras/releases/download/v0.16.0/oras_0.16.0_linux_amd64.tar.gz
# Extract it from the binary and copy it to the bin directory
tar xvzf oras.tar.gz -C /usr/local/bin oras
```

### Ratify

Install Ratify v1.0.0-beta.2 from [Ratify GitHub Release](https://github.com/deislabs/ratify/releases/tag/v1.0.0-beta.2).

```bash
# Download the Ratify binary
RATIFY_VERSION=1.0.0-beta.2
curl -Lo ratify.tar.gz https://github.com/deislabs/ratify/releases/download/v${RATIFY_VERSION}/ratify_${RATIFY_VERSION}_Linux_amd64.tar.gz
# Extract it from the binary and copy it to the bin directory
tar xvzf ratify.tar.gz -C /usr/local/bin ratify
```

## Presets

### Set up ACR and Auth information

```bash
export ACR_NAME=<YOUR_ACR_NAME>
export REGISTRY=$ACR_NAME.azurecr.io
export REPO=${REGISTRY}/net-monitor
export IMAGE=${REPO}:v1
export RESOURCE_GROUP=$ACR_NAME-acr
export LOCATION=westus3

# Create an ACR
# Premium to use tokens
az group create -n $RESOURCE_GROUP -l $LOCATION
az acr create -n $ACR_NAME -g $RESOURCE_GROUP --sku Premium
az configure --default acr=$ACR_NAME
az acr update --anonymous-pull-enabled true

# Using ACR Auth with Tokens
export NOTATION_USERNAME='wabbitnetworks-token'
export NOTATION_PASSWORD=$(az acr token create -n $NOTATION_USERNAME \
                    -r $ACR_NAME \
                    --scope-map _repositories_admin \
                    --only-show-errors \
                    -o json | jq -r ".credentials.passwords[0].value")

docker login $REGISTRY -u $NOTATION_USERNAME -p $NOTATION_PASSWORD
oras login $REGISTRY -u $NOTATION_USERNAME -p $NOTATION_PASSWORD
```

## Demo 1: Discover & Verify Signatures using Ratify

### Sign the image using ```notation```

1. We will build, push, sign the image in ACR

    ```bash
    # Build, push, sign the image in ACR
    echo $IMAGE
    ```

2. Build and push the image

    ```bash
    # build the image
    docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main

    # push the image
    docker push $IMAGE
    ```

3. Generate a test certificate

    ```bash
    # Generate a test certificate
    notation cert generate-test --default "wabbit-networks.io"
    ```

4. Sign the image

    ```bash
    notation sign $IMAGE
    ```

5. List the signatures with notation

    ```bash
    # List the signatures
    notation list $IMAGE
    ```

    > You can repeat step 4-5 to create multiple signatures to the image.

### Discover & Verify using Ratify

1. Create a Ratify config with ORAS as the signature store and notary v2 as the signature verifier.
Trust Policy reference: https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#trust-policy

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
                    "application/vnd.cncf.notary.signature": "any"
                }
            }
        },
        "verifier": {
            "version": "1.0.0",
            "plugins": [
                {
                    "name":"notaryv2",
                    "artifactTypes" : "application/vnd.cncf.notary.signature",
                    "verificationCerts": [
                        "~/.config/notation/truststore"
                    ],
                    "trustPolicyDoc": {
                        "version": "1.0",
                        "trustPolicies": [
                            {
                                "name": "default",
                                "registryScopes": [ "*" ],
                                "signatureVerification": {
                                    "level" : "strict" 
                                },
                                "trustStores": ["ca:certs"],
                                "trustedIdentities": ["*"]
                            }
                        ]
                    }
                }
            ]
        }
    }
    EOF
    ```

2. Discover the signatures

    ```bash
    # Query for the signatures
    export IMAGE_DIGEST_REF=$(docker image inspect $IMAGE | jq -r '.[0].RepoDigests[0]')
    ratify discover -s $IMAGE_DIGEST_REF
    ```

3. Verify all signatures for the image

    ```bash
    # Verify signatures
    ratify verify -s $IMAGE_DIGEST_REF
    ```

## Demo 2: Discover & Verify SBOMs, scan results using Ratify

### Generate, Sign, Push SBOMs, Scan results

1. Generate a sample SBOM

    ```bash
    echo '{"version": "0.0.0.0", "artifact": "'${IMAGE}'", "contents": "good"}' > sbom.json
    ```

2. Push the SBOM

    ```bash
    oras attach $IMAGE \
        --artifact-type org.example.sbom.v0 \
        -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
        sbom.json:application/json
    ```

3. Sign the SBOM

    ```bash
    # Capture the digest, to sign it
    SBOM_DIGEST=$(oras discover -o json \
                    --artifact-type org.example.sbom.v0 \
                    -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
                    $IMAGE | jq -r ".manifests[0].digest")

    notation sign $REPO@$SBOM_DIGEST
    ```

### Discover & Verify SBOMs and Signature using Ratify

1. Extract the SBOM plugin from the Ratify tarball and copy it to the default plugins directory

    ```bash
    tar xvf ratify.tar.gz -C ~/.ratify/plugins/ sbom
    ```

2. Create a Ratify config with ORAS as the store for SBoMs, Scan results and their corresponding signatures. Also, plugin the verifier for SBOM and scan results in the config.

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
                    "application/vnd.cncf.notary.signature": "all"
                }
            }
        },
        "verifier": {
            "version": "1.0.0",
            "plugins": [
                {
                    "name":"notaryv2",
                    "artifactTypes" : "application/vnd.cncf.notary.signature",
                    "verificationCerts": [
                        "~/.config/notation/truststore"
                    ],
                    "trustPolicyDoc": {
                        "version": "1.0",
                        "trustPolicies": [
                            {
                            "name": "default",
                            "registryScopes": [
                                "*"
                            ],
                            "signatureVerification": {
                                "level": "strict"
                            },
                            "trustStores": [
                                "ca:certs"
                            ],
                            "trustedIdentities": [
                                "*"
                            ]
                        }]
                    }
                },
                {
                    "name":"sbom",
                    "artifactTypes" : "org.example.sbom.v0",
                    "nestedReferences": "application/vnd.cncf.notary.signature"
                }
            ]
        }
    }
    EOF
    ```

3. Discover the signatures

    ```bash
    # Discover full graph of supply chain content
    ratify discover -s $IMAGE_DIGEST_REF
    ```

4. Verify the full graph of supply chain content

    ```bash
    # Verify full graph
    ratify verify -s $IMAGE_DIGEST_REF
    ```
