# Executor

The executor is the 'glue' that links all Ratify plugin-based components such as the verifiers, referrer stores, and policy providers. The executor handles all components of the verification process once a subject verification request is received either via CLI or server.


## Execution Steps

- Executor's `VerifySubject` function is invoked with a string reference to a subject to be verified.
- Subject string is parsed and extracted into a subject reference
- Find the subject descriptor associated with reference
    - Check each of the referrer stores configured and invoke `GetSubjectDescriptor` method until descriptor is successfully resolved and returned
- Iterate through each referrer store configured
    - For each store, get the list of reference descriptors for the subject from that store. (use continuation token if provided to retrieve all references)
    - Iterate through each reference descriptor found
        - Use policy provider's `VerifyNeeded` method to determine if subject with reference artifact should be verified based on configured policy
        - If verification is required, perform verification:
            - For each of the configured verifiers, check if the verifier can verify the reference artifact type
                - Invoke the verifier plugin's `Verify` method to perform verification
                - Return the `VerifyResult` containing `VerifierReport`, which has metadata such as the subject reference, verifier name, artifact type, and success status for the report. 
        - add the verifier reports returned by subject verification to a list of reports
        - if the verification result for that reference artifact was false, invoke the policy provider to determine if executor should continue to verify subsequent reference artifacts.
- After iterating through all stores, if there are no verifier reports generated, then we assume we failed to retrieve verifiable reference artifacts and error.
- Invoke the policy provider to determine the final overall success to return.
- Return a `VerifyResult` with final determined outcome from policy provider and the list of verifier reports.

## Executor Configuration
Ratify embeds OPA engine inside the executor to provide a built-in policy provider. The executor is configured using the `executor` section of the configuration. Check [`Policy Providers`](./providers.md#rego-policy-provider) section.

```json
{
    ...
    "executor": {
        "useRegoPolicy": true,
        "executionMode": "passthrough"
    },
    ...
}
```
By default, `helm install` will configure the executor to use the original config policy provider and set the execution mode to `""` as `passthrough` mode is only supported by Rego policy provider.

If you want to try use Rego policy, just set `useRegoPolicy` to `true` while deploying Ratify. And you can also set `executionMode` to `passthrough` to enable the `passthrough` mode. In this mode, Ratify will NOT make the decision but only pass the verification results to Gatekeeper for making the decision. Note that verification results returned while switching Rego policy/config policy are a bit different.

Example results when using config policy provider:
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
      "nestedResults": [
        {
          "subject": "registry:5000/sbom@sha256:a59b9a5ee8ce41fed4be7f6b8d8619bd9e619bbda6b7b1feb591c3c85f6ab7af",
          "isSuccess": true,
          "name": "notaryv2",
          "message": "signature verification success",
          "extensions": {
            "Issuer": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US",
            "SN": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US"
          },
          "artifactType": "application/vnd.cncf.notary.signature"
        }
      ],
      "artifactType": "org.example.sbom.v0"
    }
  ]
}
```

Example results when using Rego policy provider:
```json
{
  "isSuccess": true,
  "nestedReports": [
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
              "name": "notaryv2",
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