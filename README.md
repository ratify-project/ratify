# Ratify

The project provides a framework to integrate scenarios that require
verification of reference artifacts and provides a set of interfaces
that can be consumed by various systems that can participate in
artifact ratification.

**WARNING:** This is experimental code. It is not considered production-grade
by its developers, nor is it "supported" software.

## Table of Contents

- [Quick Start](#quick-start)
- [Documents](#documents)
- [Code of Conduct](#code-of-conduct)
- [Release Management](#release-management)
- [Licensing](#licensing)
- [Trademark](#trademark)

## Quick Start

### Setup

- Build the `ratify` binary

```bash
git clone https://github.com/deislabs/ratify.git
go build -o ~/bin ./cmd/ratify
```

- Build the `ratify` plugins and install them in the home directory

```bash
go build -o ~/.ratify/plugins/ ./plugins/verifier/sbom
```

- `ratify` is ready to use

```bash
$ ratify --help
Ratify is a reference artifact tool for managing and verifying reference artifacts

Usage:
  ratify [flags]
  ratify [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  discover    Discover referrers for a subject
  help        Help about any command
  referrer    Discover referrers for a subject
  serve       Run ratify as a server
  verify      Verify a subject

Flags:
  -h, --help   help for ratify

Use "ratify [command] --help" for more information about a command.
```

### Verify a Graph of Supply Chain Content

To get started with `ratify`, follow the step below.

- Create a graph of Supply Chain Content
- Discover the graph using `ratify`
- Verify the graph using `ratify`

This section outlines instructions for each of the above steps.

#### **Create a graph of Supply Chain Content**

A graph of supply chain content can be created with different tools that can manage individual supply chain objects within the graph. For this quick start, the steps outlined in [Notary V2 project] (https://notaryproject.dev/blog/2021/announcing-notation-alpha1/) will be used to create a sample graph with [`notation`](https://github.com/notaryproject/notation) and [`ORAS`](https://github.com/oras-project/oras/releases/tag/v0.2.1-alpha.1) CLI.

- Run a local instance of the [CNCF Distribution Registry](https://github.com/oras-project/distribution), with [ORAS Artifacts](https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md) support.

```bash
export PORT=5000
export REGISTRY=localhost:${PORT}
export REPO=net-monitor
export IMAGE=${REGISTRY}/${REPO}:v1

docker run -d -p ${PORT}:5000 ghcr.io/oras-project/registry:v0.0.3-alpha
```

- Build & Push an image

```bash
docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main

docker push $IMAGE
```

- Sign the image and push the signature using `notation`

registry.

```bash
notation cert generate-test --default "wabbit-networks.io"
notation sign --plain-http $IMAGE
```

- Generate a sample SBoM and push to registry

```bash
# Simulate an SBOM
echo '{"version": "0.0.0.0", "artifact": "'${IMAGE}'", "contents": "good"}' > sbom.json

# Push to the registry with the oras cli
oras push ${REGISTRY}/${REPO} \
  --artifact-type sbom/example \
  --subject $IMAGE \
  --plain-http \
  sbom.json:application/json
```

- Sign the SBoM and push the signature using `notation`

```bash
# Capture the digest of the SBOM, to sign it
SBOM_DIGEST=$(oras discover -o json \
                --artifact-type sbom/example \
                $IMAGE | jq -r ".references[0].digest")

notation sign --plain-http $REGISTRY/$REPO@$SBOM_DIGEST
```

This completes the creation of the supply chain graph.

#### **Create config with signature and SBoM verifiers**

```bash
cat <<EOF > ~/.ratify/config.json 
{ 
    "stores": { 
        "version": "1.0.0", 
        "plugins": [ 
            { 
                "name": "oras"
            }
        ]
    },
    "policies": {
        "version": "1.0.0",
        "artifactVerificationPolicies": {
            "application/vnd.cncf.notary.v2.signature": "any",
            "sbom/example": "all"
        }
    },
    "verifiers": {
        "version": "1.0.0",
        "plugins": [
            {
                "name":"notaryv2",
                "artifactTypes" : "application/vnd.cncf.notary.v2.signature",
                "verificationCerts": [
                    "/home/<user>/.config/notation/certificate/wabbit-networks-dev.crt"
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

#### Discover the graph

> Please make sure that the subject is referenced with `digest` rather
than with the tag.

```bash
export IMAGE_DIGEST_REF=$(docker image inspect $IMAGE | jq -r '.[0].RepoDigests[0]')

# Discover the graph
ratify discover -s $IMAGE_DIGEST_REF
```

#### Verify the graph

```bash
# Verify the graph
ratify verify -s $IMAGE_DIGEST_REF
```

## Documents

The [docs](docs/README.md) folder contains the beginnings of a formal
specification for the Reference Artifact Verification framework and its plugin model.

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/).

For more information see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.

## Release Management

The Ratify release process is defined in [RELEASES.md](./RELEASES.md).

## Licensing

This project is released under the [MIT License](./LICENSE).

## Trademark

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft trademarks or logos is subject to and must follow [Microsoft's Trademark & Brand Guidelines][microsoft-trademark]. Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party's policies.

[microsoft-trademark]: https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks
