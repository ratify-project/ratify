## Cosign Verifier

This README outlines how this validation framework can be used to verify signatures generated using [cosign](https://github.com/sigstore/cosign/tree/cb0c46a429253287429868c3721c9f8693797114). The verifier is added as a plugin to the framework that uses [cosign](https://github.com/sigstore/cosign/tree/cb0c46a429253287429868c3721c9f8693797114) packages to invoke the verification of an image. Currently cosign verifier works with remote registry that can provide cosign related artifacts linked as specially formatted tag to the subject artifact. It works only with [ociregistry](../../referrerstore/ociregistry) referrer store plugin that uses the OCI registry API to discover and fetch the artifacts. 

### Fallback in OCIRegistry store
A configuration flag called ```cosign-enabled``` is introduced to the plugin configuration. If this flag is enabled, the ```ListReferrers``` API will attempt to query for the cosign signatures for a subject in addition to the references queried using ```referrers API```. All the cosign signatures are returned as the reference artifacts with the artifact type ```org.sigstore.cosign.v1``` This option will enable to verify cosign signatures against any registry including the onces that don't support the [notaryproject](https://github.com/notaryproject)'s ```referrers API```. 

IMPORTANT NOTE: Cosign signatures cannot be verified from private registries. ```cosign-enabled``` flag should not be enabled for any private registry scenario even for non-cosign signature verifications.

### Configuration
The only configuration that is needed for cosign verifier is the path to the public key that is used to verify the signature. This is specified using ```key``` property in the plugin config. Here is the sample ```ratify``` config with cosign verifier

```json
{
    "store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "ociregistry",
                "cosign-enabled": true
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
                "name":"cosign",
                "artifactTypes" : "org.sigstore.cosign.v1",
                "key": "<path to cosign.pub>"
            }
        ]
    }
}
```

## Sample
This sample shows how to verify an image that has cosign signature stored with it in the registry. Please refer to cosign for usage documentation

### Generate a keypair
```bash
$ cosign generate-key-pair
Enter password for private key:
Enter again:
Private key written to cosign.key
Public key written to cosign.pub
```

### Sign a container and store the signature in the registry

```bash
$ cosign sign -key cosign.key mnltejaswini/ratifydemo
Enter password for private key:
Pushing signature to: index.docker.io/mnltejaswini/ratifydemo:sha256-1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792.sig
```

### Query for the references for the artifact (```ratify``` works with digest of the artifact)

```bash
$ ratify referrer list -s mnltejaswini/ratifydemo@sha256:1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792

mnltejaswini/ratifydemo@sha256:1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792
└── ociregistry
    └── [org.sigstore.cosign.v1]sha256:56ebc9944872035c0fea391190348a597c646b63269d434ffd1421271aeee30a
```

### Verify the cosign references using ```ratify```

```bash
$ratify verify -s mnltejaswini/ratifydemo@sha256:1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792
{
  "isSuccess": true,
  "verifierReports": [
    {
      "Subject": "mnltejaswini/ratifydemo@sha256:1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792",
      "IsSuccess": true,
      "Name": "cosign",
      "Results": [
        "cosign verification success. valid signatures found"
      ]
    }
  ]
}
```







