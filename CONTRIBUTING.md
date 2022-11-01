# Contributing to Ratify 

Welcome! We are very happy to accept community contributions to Ratify, whether those are [Pull Requests](#pull-requests), [Plugins](#plugins), [Feature Suggestions](#feature-suggestions) or [Bug Reports](#bug-reports)! Please note that by participating in this project, you agree to abide by the [Code of Conduct](./CODE_OF_CONDUCT.md), as well as the terms of the [CLA](#cla).

## Getting Started

* If you don't already have it, you will need [go](https://golang.org/dl/) v1.16+ installed locally to build the project.
* You can work on Ratify using any platform using any editor, but you may find it quickest to get started using [VSCode](https://code.visualstudio.com/Download) with the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go).
* Fork this repo (see [this forking guide](https://guides.github.com/activities/forking/) for more information).
* Checkout the repo locally with `git clone git@github.com:{your_username}/ratify.git`.
* Build the Ratify CLI with `go build -o ./bin/ratify ./cmd/ratify` or if on Mac/Linux/WSL `make build-cli`.

## Developing

### Components

The Ratify project is composed of the following main components:

* **Ratify CLI** (`cmd/ratify`): the `ratify` CLI executable.
* **Config** (`config`): an example configuration file for `ratify`.
* **Deploy** (`deploy`): the deployment files for using ratify in a Kubernetes cluster.
* **Ratify Internals** (`pkg`): the internal modules containing the majority of the Ratify code.
* **Plugins** (`plugins`): the referrer store and verifier plugins for the Ratify Framework.

### Running the tests

* You can use the following command to run the full Ratify test suite:
    * `go test -v ./cmd/ratify` on Windows
    * `make test` on Mac/Linux/WSL

### Running the Ratify CLI

* Once built run Ratify from the bin directory `./bin/ratify` for a list of the available commands.
* For any command the `--help` argument can be passed for more information and a list of possible arguments.

### Debugging Ratify with VS Code
Ratify can run through cli command or run as a http server. Create a [launch.json](https://code.visualstudio.com/docs/editor/debugging#_launch-configurations) file in the .vscode directory, then hit F5 to debug. Note the first debug session may take a few minutes to load, subsequent session will be much faster.

Sample json for cli:
```json
{
    "version": "0.2.0",
    "configurations": [{
      "name": "Debug Ratify cli ",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/ratify",
      "args": ["verify", "-s", "ratify.azurecr.io/testimage@sha256:9515b691095051d68b4409a30c4819c98bd6f4355d5993a7487687cdc6d47cc3"]
    }]
}
```
Sample launch json for http server:
```json
{
    "version": "0.2.0",
    "configurations": [{
      "name": "Debug Ratify ",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/ratify",
      "args": ["serve", "--http", ":6001", "-c", "/home/azureuser/.ratify/config.json"]
    }]
}
```
Sample curl request to invoke Ratify endpoint:

```
curl -X POST http://127.0.0.1:6001/ratify/gatekeeper/v1/verify -H "Content-Type: application/json" -d '{"apiVersion":"externaldata.gatekeeper.sh/v1alpha1","kind":"ProviderRequest","request":{"keys":["localhost:5000/net-monitor:v1"]}}'
```
### Test local changes in the k8s cluster scenario
There are some changes that should be validated in a cluster scenario.
Follow the steps below to build and deploy a Ratify image with your private changes:

#### build an image with your local changes
```
docker build -f httpserver/Dockerfile -t yourregistry/deislabs/ratify:yourtag .
```

#### [Authenticate](https://docs.docker.com/engine/reference/commandline/login/#usage) with your registry,  and push the newly built image
```
docker push yourregistry/deislabs/ratify:yourtag
```

#### Update [values.yaml](https://github.com/deislabs/ratify/blob/main/charts/ratify/values.yaml) to pull from your registry, when reusing image tag, setting pull policy to "Always" ensures we are pull the new changes
```json
image:
  repository: yourregistry/deislabs/ratify
  tag: yourtag
  pullPolicy: Always
```

### Deploy from local helm chart
Deploy using one of the following deployments.

**Option 1**
Client and server auth disabled
```
helm install ratify ./charts/ratify --atomic
```

**Option 2**
Client auth disabled and server auth enabled using self signed certificate

1. Supply a certificate to use with Ratify (httpserver) or use the following script to create a self-signed certificate.

```
./scripts/generate-tls-cert.sh
```

2. Deploy using a certificate
```
helm install ratify ./charts/ratify \
    --set provider.auth="tls" \
    --set provider.tls.skipVerify=false \
    --set provider.tls.cabundle="$(cat certs/ca.crt | base64 | tr -d '\n\r')" \
    --set provider.tls.key="$(cat certs/tls.key)" \
    --set provider.tls.crt="$(cat certs/tls.crt)" \
    --atomic
```

**Option 3**
Client auth disabled and server auth enabled using a secret

*Note: There must be an existing secret in the 'default' namespace named 'ratify-cert-secret'.*

```
# Example secret schema
apiVersion: v1
kind: Secret
metadata:
  name: ratify-cert-secret
data:
    tls.crt: <base64 crt value>
    tls.key: <base64 key value>
```

*Note: The 'provider.tls.cabundle' must be supplied. Update the path or send in a base64 encoded value.*

```
helm install ratify ./charts/ratify \
    --set provider.auth="tls" \
    --set provider.tls.skipVerify=false \
    --set provider.tls.cabundle="$(cat certs/ca.crt | base64 | tr -d '\n\r')" \
    --atomic
```

**Option 4**
Client / Server auth enabled (mTLS)  

*Note: Ratify and Gatekeeper must be installed in the same namespace which allows Ratify access to Gatekeepers CA certificate. The Ratify certificate must have a CN and subjectAltName name which matches the namespace of Gatekeeper and Ratify. For example, if installed to the namespace 'gatekeeper-system', the CN and subjectAltName should be 'ratify.gatekeeper-system'*
```
helm install ratify ./charts/ratify \
    --namespace gatekeeper-system
    --set provider.auth="mtls" \
    --set provider.tls.skipVerify=false \
    --set provider.tls.cabundle="$(cat certs/ca.crt | base64 | tr -d '\n\r')" \
    --set provider.tls.key="$(cat certs/tls.key)" \
    --set provider.tls.crt="$(cat certs/tls.crt)" \
    --atomic
```

## Pull Requests

If you'd like to start contributing to Ratify, you can search for issues tagged as "good first issue" [here](https://github.com/deislabs/ratify/labels/good%20first%20issue).

### Plugins

If you'd like to contribute to the collection of plugins:

#### Referrer Store

* A referrer store should implement the `ReferrerStore` interface (`pkg/referrerstore/api.go`).
* This should provide a mechanism to list referrers, and retrieve manifests and blob content.
* A sample referrer store is provided at `plugins/reffererstore/sample/sample.go`.

#### Verifier

* A verifier should implement the `ReferenceVerifier` interface (`pkg/verifier/api.go`).
* This should provide a mechanism to perform verification against blobs and manifests from a referrer store.
* A sample verifier is provided at `plugins/verifier/sample/sample.go`.

## Feature Suggestions

* Please first search [Open Ratify Issues](https://github.com/deislabs/ratify/issues) before opening an issue to check whether your feature has already been suggested. If it has, feel free to add your own comments to the existing issue.
* Ensure you have included a "What?" - what your feature entails, being as specific as possible, and giving mocked-up syntax examples where possible.
* Ensure you have included a "Why?" - what the benefit of including this feature will be.

## Bug Reports

* Please first search [Open Ratify Issues](https://github.com/deislabs/ratify/issues) before opening an issue, to see if it has already been reported.
* Try to be as specific as possible, including the version of the Ratify CLI used to reproduce the issue, and any example arguments needed to reproduce it.

## CLA

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
