# Verification Response

Verification Response is the data returned for external data request calls. As we add new features, the Verification Response may change its format. To make it easy for users and other developers to work with different formats, we have implemented versioning support. This document lists all the versions of the Verification Response that have ever existed and will be updated whenever we make a new change to the format.

# Versions

## 0.1.0
### Definition
```yaml
definitions:
  VerificationResponse:
    type: object
    properties:
      isSuccess:
        type: boolean
      version:
        type: string
        description: The version of the verification report
      verifierReports:
        type: array
        items:
          $ref: '#/definitions/VerifierReport'
  VerifierReport:
    type: object
    properties:
      subject:
        type: string
        description: Digested reference of the subject that was verified
      isSuccess:
        type: boolean
        description: The result of the verification
      name:
        type: string
        description: The name of the verifier that performed the verification
      message:
        type: string
        description: The message describing the verification result
      extensions:
        type: object
        description: Any extended information about the verification result
      artifactType:
        type: string
        description: The media type of the artifact being verified
      nestedResults:
        type: array
        items:
          $ref: '#/definitions/VerifierReport'
        description: The nested verification results
```
### Example
```json
{
  "isSuccess": true,
  "verifierReports": [
    {
      "subject": "registry:5000/sbom@sha256:4139357ed163984fe8ea49eaa0b82325dfc4feda98d0f6691b24f24cd6f0591e",
      "isSuccess": true,
      "name": "sbom",
      "message": "SBOM verification success. The schema is good.",
      "extensions": {
        "created": "2023-05-08T17:11:15Z",
        "creators": [
          "Organization: acme",
          "Tool: Microsoft.SBOMTool-1.0.2"
        ],
        "licenseListVersion": ""
      },
      "artifactType": "org.example.sbom.v0",
      "nestedResults": [
        {
          "subject": "registry:5000/sbom@sha256:a59b9a5ee8ce41fed4be7f6b8d8619bd9e619bbda6b7b1feb591c3c85f6ab7af",
          "isSuccess": true,
          "name": "notation",
          "message": "signature verification success",
          "extensions": {
            "Issuer": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US",
            "SN": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US"
          },
          "artifactType": "application/vnd.cncf.notary.signature"
        }
      ]
    }
  ]
}
```

## 1.0.0
### Definition
```yaml
definitions:
  VerificationResponse:
    type: object
    properties:
      version:
        type: string
        description: The version of the verification report
      isSuccess:
        type: boolean
      verifierReports:
        type: array
        items:
          $ref: '#/definitions/VerifierReport'
  VerifierReport:
    type: object
    properties:
      subject:
        type: string
        description: Digested reference of the subject that was verified
      referenceDigest:
        type: string
        description: The digest of the artifact that was verified
      artifactType:
        type: string
        description: The media type of the artifact being verified
      verifierReports:
        type: array
        items:
          $ref: '#/definitions/InnerVerifierReport'
        description: The verification reports related to the artifact
      nestedReports:
        type: array
        items:
          $ref: '#/definitions/VerifierReport'
        description: The nested verification reports
  InnerVerifierReport:
    type: object
    properties:
      isSuccess:
        type: boolean
        description: The result of the verification
      message:
        type: string
        description: The message describing the verification result
      name:
        type: string
        description: The name of the verifier that performed the verification
      extensions:
        type: object
        description: Any extended information about the verification result
```

###
```json
{
  "version": "1.0.0",
  "isSuccess": true,
  "verifierReports": [
    {
      "subject": "registry:5000/sbom@sha256:f291201143149e4006894b2d64202a8b90416b7dcde1c8ad997b1099312af3ce",
      "referenceDigest": "sha256:932dde71a9f26ddafa61e8a7df2b296b1787bcb6e75c515584a53776e81a8a00",
      "artifactType": "org.example.sbom.v0",
      "verifierReports": [
        {
          "isSuccess": true,
          "message": "SBOM verification success. The schema is good.",
          "name": "sbom",
          "extensions": {
            "created": "2023-05-11T05:20:43Z",
            "creators": [
              "Organization: acme",
              "Tool: Microsoft.SBOMTool-1.1.0"
            ],
            "licenseListVersion": ""
          }
        }
      ],
      "nestedReports": [
        {
          "subject": "registry:5000/sbom@sha256:932dde71a9f26ddafa61e8a7df2b296b1787bcb6e75c515584a53776e81a8a00",
          "referenceDigest": "sha256:bf67213f8e048c2262b1dd007a4380f03431e1aa2ab58c7afdce7c2f763f7684",
          "artifactType": "application/vnd.cncf.notary.signature",
          "verifierReports": [
            {
              "isSuccess": true,
              "message": "signature verification success",
              "name": "notation",
              "extensions": {
                "Issuer": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US",
                "SN": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US"
              }
            }
          ],
          "nestedReports": []
        }
      ]
    }
  ]
}
```