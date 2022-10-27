# Verify a artifact graph on commandline 

## Setup

- Build the `ratify` binary and ensure that `~/bin` is in your `PATH`. 

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
  resolve     Resolve digest of a subject that is referenced by a tag
  help        Help about any command
  referrer    Discover referrers for a subject
  serve       Run ratify as a server
  verify      Verify a subject

Flags:
  -h, --help   help for ratify

Use "ratify [command] --help" for more information about a command.
```

## Verify a Graph of Supply Chain Content

To get started with `ratify`, follow the step below.

- Create a graph of Supply Chain Content
- Discover the graph using `ratify`
- Verify the graph using `ratify`

This section outlines instructions for each of the above steps.

### **Create a graph of Supply Chain Content**

A graph of supply chain content can be created with different tools that can manage individual supply chain objects within the graph. For this quick start, the steps outlined in [Notary V2 project](https://notaryproject.dev/blog/2021/announcing-notation-alpha1/) will be used to create a sample graph with [`notation`](https://github.com/notaryproject/notation) and [`ORAS`](https://github.com/oras-project/oras/releases/tag/v0.2.1-alpha.1) CLI.

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
oras attach $IMAGE \
  --artifact-type sbom/example \
  --plain-http \
  sbom.json:application/json
```

- Sign the SBoM and push the signature using `notation`

```bash
# Capture the digest of the SBOM, to sign it
SBOM_DIGEST=$(oras discover -o json \
                --artifact-type sbom/example \
                $IMAGE | jq -r ".referrers[0].digest")

notation sign --plain-http $REGISTRY/$REPO@$SBOM_DIGEST
```

This completes the creation of the supply chain graph.

#### **Create config with signature and SBoM verifiers**

```bash
cat <<EOF > ~/.ratify/config.json 
{ 
    "store": { 
        "version": "1.0.0", 
        "plugins": [ 
            { 
                "name": "oras",
                "useHttp": true
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

### Discover the graph

> If the subject is referenced by tag, it is resolved to digest before verifying it.

```bash
# Discover the graph
ratify discover -s $IMAGE
```

### Verify the graph

```bash
# Verify the graph
ratify verify -s $IMAGE
```
