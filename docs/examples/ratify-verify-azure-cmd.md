# Notary v2 Signature Verification With ACR Using Ratify

## Install Notation, ORAS, and Ratify

### Notation

Install Notation v1.0.0-rc.1 with plugin support from [Notation GitHub Release](https://github.com/notaryproject/notation/releases/tag/v1.0.0-rc.1).

```bash
# Download the Notation binary
curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v1.0.0-rc.1/notation_1.0.0-rc.1_linux_amd64.tar.gz
# Extract it from the binary
tar xvzf notation.tar.gz
# Copy the Notation CLI to your bin directory
mkdir -p ~/bin && cp ./notation ~/bin
```

### ORAS

Install ORAS 0.16.0 on a Linux machine. You can refer to the [ORAS installation guide](https://oras.land/cli/) for details.

```bash
# Download the ORAS binary
curl -LO https://github.com/oras-project/oras/releases/download/v0.16.0/oras_0.16.0_linux_amd64.tar.gz
# Create a folder to extract the ORAS binary
mkdir -p oras-install/
tar -zxf oras_0.16.0_*.tar.gz -C oras-install/
# Copy the Notation CLI to your bin directory
mv oras-install/oras /usr/local/bin/
```

### SBOM Tool

The SBOM tool will be used to create an SPDX 2.2 compatible SBOM. you can refer to the [SBOM Tool README](https://github.com/microsoft/sbom-tool) for more details on installation and usage.

```bash
curl -Lo sbom-tool https://github.com/microsoft/sbom-tool/releases/latest/download/sbom-tool-linux-x64 
chmod +x sbom-tool 
mv sbom-tool /usr/local/bin/
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

4. Sign the image[Notation Sign command only works in the next rc.1 release version]

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
                        "~/.config/notation/truststore/x509/ca/wabbit-networks.io/wabbit-networks.io.crt"
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

1. Clone the Ratify Repo. As a sample, it will be used to generate the SBOM.

    ```bash
    git clone https://github.com/deislabs/ratify.git
    cd ratify
    ```

2. Generate an SBOM

    See [sbom-generation](https://github.com/microsoft/sbom-tool#sbom-generation) for more information on how to generate an SBOM.

    ```bash
    # Ensure you are in the ratify directory from the previous step
    sbom-tool generate -b . -bc . -pn ratify -pv 1.0 -ps ratify-test -nsb https://github.com/deislabs -V
    ```

3. Push the SBOM

    ```bash
    oras attach $IMAGE \
        --artifact-type org.example.sbom.v0 \
        -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
        ./_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json
    ```

4. Sign the SBOM

    ```bash
    # Capture the digest, to sign it
    SBOM_DIGEST=$(oras discover -o json \
                    --artifact-type org.example.sbom.v0 \
                    -u $NOTATION_USERNAME -p $NOTATION_PASSWORD \
                    $IMAGE | jq -r ".manifests[0].digest")

    notation sign $REPO@$SBOM_DIGEST
    ```

### Discover & Verify SBOMs and Signature using Ratify

1. Build the SBOM plugin

    Ensure you have cloned the Ratify repo and are in the ratify directory from step 1 in the previous section.

    ```bash
    go build ./plugins/verifier/sbom/
    mv sbom ~/.ratify/plugins/
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
                        "~/.config/notation/localkeys/wabbit-networks.io.crt"
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
