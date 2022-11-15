# Executor

The executor is the 'glue' that links all Ratify plugin-based components such as the verifiers, referrer stores, and policy providers. The executor handles all components of the verification process once a subject verification request is received either via CLI or server.

## Executor Configuration format

The following are the keys used to describe configuration of the executor.

| Property | Type | IsRequired | Description |
| -------- | ---- | ---------- | ----------- |
| requestTimeout    | int (milliseconds) | false; Default: `2800`; Helm chart default: `6800`;| The timeout for the verification request. |
| nestedReferences  | object[string]bool | false | The key is an artifact type. A value of true will initiate nested verification for all references with that artifact type |

### Nested References Configuration object

The `nestedReferences` config block MAY be used to verify nested references on a per artifact type basis. Setting an artifact type to `true` will initiate [nested verification](https://hackmd.io/9htAyk-OQmauWPnNqMTVIw?both#Nested-Verification) for any reference with that artifact type. Nested references for omitted artifacts types will not be retrieved or verified. 

>note: If there is no verifier for an artifact type supplied in the `nestedReferences` config, the executor's `VerifySubject` method will be called for that reference but the verification will be skipped since no verifier is configured to verify that artifact type.

#### JSON Config Example

Verify nested references for all references with an artifact type of `org.example.sbom.v0`

```json
{
    "executor": {
        "nestedReferences":{
            "org.example.sbom.v0": true
        }
    },
}
```

#### Helm Chart Values Example

Verify nested references for all references with an artifact type of `org.example.sbom.v0`

```yaml
executor:
    nestedReferences:
        org.example.sbom.v0: true
```

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
                - If the reference artifact type is [configured for nested references](./executor.md#nested-references-configuration-object), the reference artifact will be verified by invoking `executor.VerifiySubject()`. The results will be appended to the reference's `VerifierReport` as `NestedResults`
        - add the verifier reports returned by subject verification to a list of reports
        - if the verification result for that reference artifact was false, invoke the policy provider to determine if executor should continue to verify subsequent reference artifacts.
- After iterating through all stores, if there are no verifier reports generated, then we assume we failed to retrieve verifiable reference artifacts and error.
- Invoke the policy provider to determine the final overall success to return.
- Return a `VerifyResult` with final determined outcome from policy provider and the list of verifier reports.
