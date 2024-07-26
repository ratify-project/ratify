# Error messages improvements

## Problem/Motivation

Error messages are crucial because they provide specific information about what went wrong, helping you quickly identify and resolve issues. Detailed error messages can pinpoint the exact part of the configuration or code that caused the problem. Error messages can include links or references to documentation, guiding you on how to fix the issue promptly.

While testing the Ratify policy for signature verification, we noticed that some error messages were difficult to comprehend from the user perspective.

Error message example 1:

```text
verification failed: Error: referrers not found, Code: REFERRERS_NOT_FOUND, Component Type: executor`
```

The error message is brief but lacks clarity regarding the cause and required user actions. It's also confusing as to what the term "referrer" signifies.

Error message example 2:

```text
"Original Error: (Original Error: (signature is not produced by a trusted signer), Error: verify signature failure, Code: VERIFY_SIGNATURE_FAILURE, Plugin Name: notation, Component Type: verifier, Documentation: https://github.com/notaryproject/notaryproject/tree/main/specs, Detail: failed to verify signature of digest), Error: verify reference failure, Code: VERIFY_REFERENCE_FAILURE, Plugin Name: notation, Component Type: verifier"
```

This error message is lengthy and contains nested "Original Error" details. It explains the cause of the error and offers directions to resolve it via a linked document, but its structure is poor.

Error message example 3:

```text
"verifierReports": [
  {
        subject: "docker.io/library/hello-world@sha256:1408fec50309afee38f3535383f5b09419e6dc0925bc69891e79d84cc4cdce6",
        "isSuccess": false,
        "message: "verification failed: Error: no verifier report, Code: NO_VERIFIER _REPORT, Component Type: executor, Description: No verifier report was generatec preventing access to the registry, or the absence of appropriate verifiers corresponding to the referenced image artifacts."
  }
]
```

The json object "verifierReports" appears in the error logs. Its format is organized and helpful for policy engines dealing with error information as needed. Yet, the message field is similar to other error messages and lacks clarity on the problem, its cause, and remediation methods. We also observed that it contains different supported properties when contrasted with the Ratify config policy and Rego policy, which is inconsistent.

Additional results documented at https://hackmd.io/tewiH4DvST2zcgXJNeLLyw#Ratify-Error-Handling-Scenarios span various scenarios, including KMP, Store, Verifier, Policy configuration, access control, and signature verification errors.

The document aims to provide solutions and guidelines to improve error messages in different scenarios.

## Scenarios

## Proposed solutions

## Error messages guidelines
