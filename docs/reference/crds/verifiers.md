Ratify supports many verifiers to validate different artifact types. View more CRD samples [here](../../../config/samples/). Each verifier must specify the `name` of the verifier and the `artifactType` this verifier handles. Common properties:

```yml
name: required, name of the verifier
artifactType: required, the type of artifact this verifier handles
address: optional. Plugin path, defaults to value of env "RATIFY_CONFIG" or "~/.ratify/plugins"
source:  optional. Source location to download the plugin binary
  artifact:  e.g. wabbitnetworks.azurecr.io/test/sample-verifier-plugin:v1
parameters: optional. Parameters specific to this verifier
```
 

## Notary

Sample Notary yaml spec:
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Verifier
metadata:
  name: verifier-notary
spec:
  name: notaryv2
  artifactTypes: application/vnd.cncf.notary.signature
  parameters:
    verificationCertStores:  # certificates for validating signatures
      certs: # name of the trustStore
        - certstore-akv # name of the certificate store CRD to include in this trustStore
        - certstore-akv1 
    trustPolicyDoc: # policy language that indicates which identities are trusted to produce artifacts
      version: "1.0"
      trustPolicies:
        - name: default
          registryScopes:
            - "*"
          signatureVerification:
            level: strict
          trustStores:
            - ca:certs
          trustedIdentities:
            - "*"
```

| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| verificationCerts      | no    |      An array of string. Notary verifier will load all certificates from path specified in this array        |   ""            |
| verificationCertStores      | no    |    Defines a collection of cert store objects. This property supersedes the path defined in `verificationCerts`      |       ""        |
| trustPolicyDoc   | yes     |   [Trust policy](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md) is a policy language that indicates which identities are trusted to produce artifacts.          |     ""    |

## Cosign
Cosign verifier can be used to verify signatures generated using [cosign](https://github.com/sigstore/cosign/), learn more about the plugin [here](../../../plugins/verifier/cosign/README.md)
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Verifier
metadata:
  name: verifier-cosign
spec:
  name: cosign
  artifactTypes: org.sigstore.cosign.v1
  parameters:
    key: /usr/local/ratify-certs/cosign/cosign.pub
```
| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| key      | Yes    |     Path to the public key used for validating the signature    |   ""            |

## Sbom
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Verifier
metadata:
  name: verifier-sbom
spec:
  name: sbom
  artifactTypes: org.example.sbom.v0
  parameters: 
    nestedReferences: application/vnd.cncf.notary.signature
```
| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| nestedReferences      | No    | If nestedReferences contains any string value, the nested artifact will also be verified. For example, customer might want to enforce notary signatures on all sboms.|   ""            |

## Schemavalidator

Validate Json artifacts against JSON schemas, learn more about the plugin [here](../../../plugins/verifier/schemavalidator/README.md)

```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Verifier
metadata:
  name: verifier-schemavalidator
spec:
  name: schemavalidator
  artifactTypes: vnd.aquasecurity.trivy.report.sarif.v1
  parameters: 
    schemas:
      application/sarif+json: https://json.schemastore.org/sarif-2.1.0-rtm.5.json
```
| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| schemas      | Yes    |     A mapping between the schema name to the schema path. The path can be either a URL or a canonical file path that start with file:// |   ""            |
