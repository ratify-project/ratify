# Hora

The project provides a framework to integrate scenarios that require
verification of references and aim to provide a set of interfaces that can
be consumed by various systems that can participate in artifact verification.

**WARNING:** This is experimental code. It is not considered production-grade
by its developers, nor is it "supported" software.

## The Reference Artifact Verifier Specification

The [docs](docs/README.md) folder contains the beginnings of a formal
specification for the Reference Artifact Verification toolset.

## Licensing

This project is released under the [MIT License](./LICENSE).

### Trademark

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft trademarks or logos is subject to and must follow [Microsoft’s Trademark & Brand Guidelines][microsoft-trademark]. Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party’s policies.

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/).

For more information see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.

## Setup & Usage

- Build the ```hora``` binary

```bash
git clone https://github.com/deislabs/hora.git
git checkout dev
go build -o ~/bin ./cmd/hora
```

- Build the ```hora``` plugins and install them in the home directory

```bash
go build -o ~/.hora/plugins ./plugins/verifier/sbom
```

- Update the ```./config/config.json``` to the certs folder and copy it to home dir

```bash
cp ./config/config.json ~/.hora
```

- ```hora``` is ready to use

```bash
Usage:
  hora [flags]
  hora [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  referrer    Discover referrers for a subject
  verify      Verify a subject

Flags:
  -h, --help   help for hora

Use "hora [command] --help" for more information about a command.
```

## Try it Out
- Download the [notation CLI](https://github.com/notaryproject/notation/releases/tag/v0.7.0-alpha.1)
- Pull and build the [oras CLI](https://github.com/oras-project/oras/tree/artifacts) (ensure you build from the artifacts branch)
- Run a local registry with oras support:

```shell
docker run -d -p 5000:5000 ghcr.io/oras-project/registry:v0.0.3-alpha
```

- Build a docker image to test with:

```shell
docker build -t localhost:5000/net-monitor:v1 https://github.com/wabbit-networks/net-monitor.git#main
```

- Push the image to the local registry:

```shell
docker push localhost:5000/net-monitor:v1
```

- Use notation to generate a test certificate:

```shell
notation cert generate-test --default "wabbit-networks.io"
```

- Sign the image using notation:

```shell
notation sign --plain-http localhost:5000/net-monitor:v1
```

- Create a test SBoM (we set contents: bad to simulate a failure during verification):

```shell
echo '{"version": "0.0.0.0", "artifact": "net-monitor:v1", "contents": "bad"}' > sbom.json
```

- Push the SBoM using oras:

```shell
oras push localhost:5000/net-monitor \
--artifact-type application/x.example.sbom.v0 \
--subject localhost:5000/net-monitor:v1 \
--export-manifest sbom-manifest.json \
--plain-http \
./sbom.json:application/json
```

- Verify both artifacts above exist using hora (WHEN THIS WORKS, CURRENTLY WIP):

```shell
hora referrer list \
-c ./hora/config/config.json \
-s $(docker image inspect localhost:5000/net-monitor:v1 | jq -r '.[0].RepoDigests[0]')

example output
sha256:bdad7c3a3209b464c0fdfcaac4a254f87448bc6877c8fd2a651891efb596b05a
└── oras 
    ├── [application/x.example.sbom.v0]sha256:110b9d8d880ea0a0ebb3df590faabf239fda1a80d6b64b38dc9ad9cf29aeca5f
    │   └── [application/vnd.cncf.notary.v2]sha256:f71ad4bf25ec8ed0cfd60b22b895f90264fa8a7e8ea62b8ad72f8616d9102d67
    └── [application/vnd.cncf.notary.v2]sha256:b95fabe5f87248540a5af1cd194841a322548ef46144e6d085d3cca00cc843a8
```

- You can also see artifacts using oras:

```shell
oras discover --plain-http localhost:5000/net-monitor:v1
```

- ```hora``` can be used to verify all the references to the target image.
Please make sure that the image is referenced with ```digest``` rather
than with the tag.

```json
hora verify -s $(docker image inspect localhost:5000/net-monitor:v1 | jq -r '.[0].RepoDigests[0]')

{
  "isSuccess": true,
  "verifierReports": [
    {
      "IsSuccess": false,
      "Name": "sbom",
      "Results": [
        "SBOM verification completed. contents bad"
      ]
    },
    {
      "IsSuccess": true,
      "Name": "nv2verifier",
      "Results": [
        "Notary verification success"
      ]
    }
  ]
}
```

In the above sample, the verification is still success because the
[policy- ContinueVerifyOnFailure](./pkg/policyprovider/configpolicy/configpolicy.go)
is set to ```true```. If it is set to false, the verification will be stopped at the first failure.


[microsoft-trademark]: https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks