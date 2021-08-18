## Licensing
This project is released under the [MIT License](./LICENSE).

## Code of Conduct

This project has adopted the [Microsoft Open Source Code of
Conduct](https://opensource.microsoft.com/codeofconduct/).

For more information see the [Code of Conduct
FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact
[opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional
questions or comments.

## Hora

## Setup & Usage

- Build the ```hora``` binary 

```
git clone https://github.com/deislabs/hora.git
git checkout dev
go build -o ~/bin ./cmd/hora
```
- Build the ```hora``` plugins and install them in the home directory

```
go build -o ~/.hora/plugins ./plugins/referrerstore/ociregistry
go build -o ~/.hora/plugins ./plugins/verifier/nv2verifier
go build -o ~/.hora/plugins ./plugins/verifier/sbom
```

- Update the ```./config/config.json``` to the certs folder and copy it to home dir

```
cp ./config/config.json ~/.hora
```

- ```hora``` is ready to use
```
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

- Follow [nv2 demo-script](https://github.com/notaryproject/nv2/blob/prototype-2/docs/nv2/demo-script.md) to setup local registry, push an image with a signature and push a signed SBOM to the image. 
- In the script above, for testing purpose, push a SBOM with ```content:bad``` to simulate a failed verification
```
echo '{"version": "0.0.0.0", "artifact": "net-monitor:v1", "contents": "bad"}' > sbom.json
```

- ```hora``` can be used to verify all the references to the target image. Please make sure that the image is referenced with ```digest``` rather than with the tag.

```
hora verify -s $IMAGE_DIGEST

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
In the above sample, the verification is still success because the [policy- ContinueVerifyOnFailure](./pkg/policyprovider/configpolicy/configpolicy.go) is set to ```true```. If it is set to false, the verification will be stopped at the first failure. 

- ```hora``` can also be used to query for the references

```
hora referrer list -s $IMAGE_DIGEST
```
that will generate an output like below 

```
sha256:bdad7c3a3209b464c0fdfcaac4a254f87448bc6877c8fd2a651891efb596b05a
└── ociregistry
    ├── [application/x.example.sbom.v0]sha256:110b9d8d880ea0a0ebb3df590faabf239fda1a80d6b64b38dc9ad9cf29aeca5f
    │   └── [application/vnd.cncf.notary.v2]sha256:f71ad4bf25ec8ed0cfd60b22b895f90264fa8a7e8ea62b8ad72f8616d9102d67
    └── [application/vnd.cncf.notary.v2]sha256:b95fabe5f87248540a5af1cd194841a322548ef46144e6d085d3cca00cc843a8
```
