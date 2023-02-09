Ratify supports many verifiers to validate different artifact types. View more CRD samples [here](../../../config/samples/). Each verifier must specify the name of the verifier and the artifact type this verifier handles. Common properties:

```yml
name: required, name of the verifier
artifactType: required, the type of artifact this verifier handles
address: optional. Plugin path, defaults to value of env "RATIFY_CONFIG" or "~/.ratify/plugins"
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
| verificationCertStores      | no    |    verificationCertStores property defines a collection of cert store objects. This property supersedes path defined in verificationCerts      |       ""        |
| trustPolicyDoc   | yes     |   [Trust policy](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md) is a policy language that indicates which identities are trusted to produce artifacts.          |     ""    |

## Cosign
Doc coming soon..

## Sbom
Doc coming soon..

## Schemavalidator
Doc coming soon..

## Dynamic
Doc coming soon..