# Error messages improvements

## Problem/Motivation

Error messages are crucial because they provide specific information about what went wrong, helping users quickly identify and resolve issues. Detailed error messages can pinpoint the exact part of the configuration or code that caused the problem. Error messages can include links or references to documentation, guiding users on how to fix the issue promptly. While testing the Ratify policy for signature verification, we noticed that some error messages were difficult to comprehend from the user perspective.

Error message example 1:

```text
time=2024-07-17T16:28:16.939576441Z level=warning msg=Original Error: (Original Error: (HEAD "https:/
roacr.azurecr.io/v2/net-monitor/manifests/v2": GET "https://roacr.azurecr.io/oauth2/token?
scope=repository%3Anet-monitor%3Apull&service=roacr.azurecr.io": response status code 401: unauthorized: 
authentication required, visit https://aka.ms/acr/authorization for more information.), Error: repository 
operation failure, Code: REPOSITORY_OPERATION_FAILURE, Plugin Name: oras), Error: get subject descriptor 
failure, Code: GET_SUBJECT_DESCRIPTOR_FAILURE, Plugin Name: oras, Component Type: referrerStore, Detail: 
failed to resolve the subject descriptor component-type=referrerStore go.version=go1.21.10 namespace= 
trace-id=34b27888-5402-443e-9836-77124c840561
```

The above example indicated the an error happened, however, a warning level was set for the log. The error message was set to the field `msg`. It contained nested errors. The first original error message correctly described the source of the problem "401 unauthorized" and pointed to a document for resolution, however, errors following the original error were redundant and not well formatted, thus complicated the overall message. The overall message failed to describe the context of the error, although "401 unauthorized" explained the reason, but did this error happen during signature verification or else? What does the `subject` mean?

Error message example 2:

```text
"verifierReports": [
  {
        "subject": "docker.io/library/hello-world@sha256:1408fec50309afee38f3535383f5b09419e6dc0925bc69891e79d84cc4cdce6",
        "isSuccess": false,
        "message": "verification failed: Error: no verifier report, Code: NO_VERIFIER _REPORT, Component Type: 
        executor, Description: No verifier report was generatec preventing access to the registry, or the absence 
        of appropriate verifiers corresponding to the referenced image artifacts."
  }
]
```

When Ratify completes artifact verification, the result is returned to the policy engine in the format of the json object `verifierReports`. The `verifierReports` is also recorded in an INFO log of Ratify. If `isSuccess` field is set to `false`, the `message` field is set to error messages. In above example, the message is not well formatted and lacked clarity on the problem, its cause, and remediation methods. It's hard for users to understand in what context the error happened and what users need to do. For example, is it an error for signature verification? What does "no verifier report" mean? We also observed that the `verifierReports` contains different supported fields when compared with the Ratify Config policy and Ratify Rego policy, which is inconsistent.

Error message example 3:

```text
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: 
[ratify-constraint] Subject failed verification: huishwabbit1.azurecr.io/
test8may24@sha256:c780036bc8a6f577910bf01151013aaa18e255057a1653c76d8f3572aa3f6ff6
```

The policy engine, for instance, Gatekeeper, has produced the above error message using a constraint template supplied by Ratify. It is the responsibility of the policy engine to tailor the constraint template for proper error messages to their requirements, however, it requires Ratify to provided useful verification reports as data inputs. In this example, the error message is not clear to users regarding the meaning of the term `Subject`, and fails to specify the context of failure, such as whether it was related to signature verification or SBOM verification or other verifications. Additionally, reasons behind the error were not provided. Furthermore, users may not be able to locate this error in the complete K8s logs to view more logs during error happened, because only artifact digest was shown and it is not enough to pinpoint the exact error in K8s logs.

Further findings covering a range of cases such as Key Management Provider (KMP), Store, Verifier, Policy configuration, access control, and signature verification issues are recorded at [Ratify Error Handling Scenarios.md](../discussion/Ratify%20Error%20Handling%20Scenarios.md)

In summary, the areas that need enhancement include:

- Error messages similar to the first example will appear in Ratify logs. These may be found in the logs of Ratify Pods if Ratify is set up as a Kubernetes service, or they might be output by the Ratify CLI. The primary concerns include excessive nested errors, lacking error context, and not user-friendly error descriptions.
- Error messages contained within `verifierReports` that are sent back to the policy engine, as seen in the second example. These error messages share issues similar to those mentioned for the first example. The policy engine can customize their Rego policies using the messages inside the `verifierReports` to return to users in the logs or UI of the policy engine.
- Ratify failed to provide sufficient information for the policy engine to generate error messages, which is demonstrated in the third example.

The document aims to provide solutions and guidelines to improve error messages.

## Scenarios

### Error messages displayed in the Ratify logs

Alice works as a DevOps engineer at Contoso. She set up tasks to deploy containerized apps into Kubernetes clusters. The cluster is assigned with the policy to deny the deployment of images that don't pass policy evaluation including verification of signature, SBOM, vulnerability reports and other image metadata. Alice knows that behind the scene, it is the Ratify conduct the verification and returned results as reports to the policy engine. When policy evaluation fails, Alice sees clear and actionable error messages in Ratify logs. The error messages contain concise error descriptions, error reasons and error recommendations, allowing her to act on errors promptly.

### Error messages displayed in verification reports used by the policy engine

Bob is a software engineer on Contoso's Policy team, writing policies used during admissions in Kubernetes clusters. These policies evaluate images based on verifier reports generated by Ratify. If policy evaluation fails, Ratify sends back the reports with error messages to the policy engine. The reports, in JSON format, provide structured error messages that Bob utilizes to create clear and actionable error messages for the policy engine. These messages include concise error descriptions, error reasons, and error recommendations, allowing policy users to act on errors promptly.

### Error messages returned by Ratify CLI commands

Gina is a software engineer on the CI/CD team at Contoso, where she creates pipeline tasks incorporating Ratify CLI commands to assess artifacts according to policies. Should a policy check not pass, the corresponding artifacts are prevented from progressing in the pipeline. When policy evaluation fails, Gina sees concise, clear and actionable error messages returned by Ratify CLI commands. The error messages contain concise error descriptions, error reasons and error recommendations, allowing her to act on errors promptly.

## Proposed solutions

We won’t create new error message guidelines; instead, we’ll refer to the existing ones. [Azure CLI Error Handling Guidelines](https://github.com/Azure/azure-cli/blob/dev/doc/error_handling_guidelines.md#error-message) outlined a general pattern for error messages, consisting of:

1. __What the error is.__
2. __Why it happens.__
3. __What users need to do to fix it.__

The proposed improvements for Ratify error messages adhere to this general pattern and the detailed DOs and DON'Ts provided in the guidelines. The error message will also include an error code. Since Ratify already supports a list of error codes, these can be used to search for remediation in the troubleshooting guide. For example, search error code `CERT_INVALID` in [the troubleshooting guide](https://ratify.dev/docs/troubleshoot/key-management-provider/kmp-tsg#cert_invalid).

The recommended format for an error message in the Ratify log is as following.

```text
"<Error Code>: <Error Description>: <Error Reason>: <Remediation>"
```

For the error messages displayed in `verifierReports`, it is recommended to add two new optional fields `errorReason` and `remediation`, which will be used when the field `isSuccess` is set to `false`:

```text
"verifierReports": [
  {
        "subject": "<Digest of the Artifact>"
        "isSuccess": false,
        "message": "<Error Description>",
        "errorReason": "<Error Reason>",
        "remediation": "<Remediation>"
  }
]
```

## Examples

### Error messages displayed in the Ratify logs or returned by Ratify CLI commands

For the above first example, the error message in the Ratify log can be improved to:

```text
REPOSITORY_OPERATION_FAILURE: Failed to resolve the artifact descriptor: HEAD "https://roacr.azurecr.io/v2/net-monitor/manifests/v2": GET "https://roacr.azurecr.io/oauth2/token?
scope=repository%3Anet-monitor%3Apull&service=roacr.azurecr.io": response status code 401: unauthorized: 
authentication required, visit https://aka.ms/acr/authorization for more information.
```

### Error messages displayed in `verifierReports`

For the second example, the error message can be improved to:

```text
"verifierReports": [
  {
        "subject": "docker.io/library/hello-world@sha256:1408fec50309afee38f3535383f5b09419e6dc0925bc69891e79d84cc4cdce6",
        "isSuccess": false,
        "message": "NO_VERIFIER_REPORT: Failed to verify artifact docker.io/library/hello-world@sha256:1408fec50309afee38f3535383f5b09419e6dc0925bc69891e79d84cc4cdce6: 
        "errorReason": "No signature is found or wrong configuration"
        "remediation": "Please either sign the artifact or configure verifiers for signature verification. Learn more at https://ratify.dev/docs/plugins/verifier/notation."
  }
]
```

> This link https://ratify.dev/docs/plugins/verifier/notation is used as an example to illustrate the improvements. The link should vary depending on the particular error encountered.

## References

- [Azure CLI Error Handling Guidelines](https://github.com/Azure/azure-cli/blob/dev/doc/error_handling_guidelines.md)
- [ORAS CLI Error Handling and Message Guideline](https://github.com/oras-project/oras/blob/v1.2.0/docs/proposals/error-handling-guideline.md)
