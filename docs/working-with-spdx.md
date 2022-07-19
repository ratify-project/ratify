# Working with SPDX
SPDX is a popular specification for representing software bill of material (SBoM) information. Once an SBoM has been
generated there are a number of ways to validate the contents. This guide will walk through the steps to generate an
SBoM from a container image and validate some information from the SBoM using the example `licensechecker` plugin. By
the end of this guide you will have:

- Built a docker image
- Generated an SBoM for that image in SPDX tag-value format
- Pushed the SBoM to a registry
- Configured Ratify to validate the licenses used in the docker image
- Run the validation via the CLI

For more information on additional metadata that is captured in SPDX visit the [SPDX specification docs](https://spdx.dev/specifications/).

## Setup
In order to complete this guide you will need some tools.

- [docker](https://www.docker.com/get-started): This tool will be used to build the container image
- [syft](https://github.com/anchore/syft): This tool will be used to generate an SBoM in SPDX tag-value format
- [oras](https://github.com/oras-project/oras): This tool will be used to push the generated SBoM to the registry
- [ratify](https://github.com/deislabs/ratify): This tool will be used to validate the SBoM

You will also need a registry to push your container image and SBoM to. For this guide deploy a local registry with
oras artifacts support:
```shell
docker run --name sbom_demo -d -p 5000:5000 ghcr.io/oras-project/registry:v0.0.3-alpha
```

## Build the Image
First we need to build the container image we will be using:
```shell
docker build -t localhost:5000/only-spdx:v1 https://github.com/wabbit-networks/net-monitor.git\#main
```

## Generate the SBoM
Generate an SPDX SBoM using syft:
```shell
syft -o spdx --file sbom.spdx localhost:5000/only-spdx:v1
```

## Push the image and SBoM
After building the image and generating the SBoM push them to the registry:
```shell
docker push localhost:5000/only-spdx:v1

oras push localhost:5000/only-spdx \
  --artifact-type application/vnd.ratify.spdx.v0 \
  --subject localhost:5000/only-spdx:v1 \
  --plain-http \
  sbom.spdx:application/text
```

## Validating the SBoM
If you have not done so already, build and install ratify:
```shell
make build
make install
```

Next we will create the config file we will use for validation:
```shell
cat <<'EOF' >> spdxconfig.json
{
    "store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache"
            }
        ]
    },
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.ratify.spdx.v0": "all"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "licensechecker",
                "artifactTypes": "application/vnd.ratify.spdx.v0",
                "allowedLicenses": [
                    "GPL-2.0-only",
                    "MIT",
                    "OpenSSL",
                    "BSD-2-Clause AND BSD-3-Clause",
                    "Zlib",
                    "MPL-2.0 AND MIT",
                    "ISC"
                ]
            }

        ]

    }
}
EOF
```

This config file will only enable the `licensechecker` verifier and configure it to check that only the licenses in the
`allowedLicenses` list are used. Note here that the `artifactType` of the SBoM is arbitrary and can be whatever you
want to specify it as so long as you are consistently use it when pushing and validating.

We are now ready to validate our image:
```shell
ratify verify -c spdxconfig.json -s localhost:5000/only-spdx:v1

{
  "isSuccess": true,
  "verifierReports": [
    {
      "subject": "localhost:5000/only-spdx:v1",
      "isSuccess": true,
      "name": "licensechecker",
      "results": [
        "License Check: SUCCESS",
        "All packages have allowed licenses"
      ]
    }
  ]
}
```

Now remove one of the license lines from the config file, for example `MIT`, and run the validation again:
```shell
ratify verify -c spdxconfig.json -s localhost:5000/only-spdx:v1

{
  "isSuccess": false,
  "verifierReports": [
    {
      "subject": "localhost:5000/only-spdx:v1",
      "name": "licensechecker",
      "results": [
        "License Check: FAILED",
        "package 'alpine-keys' has unpermitted license 'MIT'",
        "package 'musl-utils' has unpermitted license 'MIT'",
        "package 'musl' has unpermitted license 'MIT'"
      ]
    }
  ]
}
```

We have successfully used Ratify to validate the contents of an SPDX SBoM!

## Cleaning Up
Remove the registry container, the guide sbom, and the guide config:
```shell
docker container stop sbom_demo
docker container rm sbom_demo

rm sbom.spdx
rm spdxconfig.json
```
