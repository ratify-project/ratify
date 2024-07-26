# Error messages improvements

## Problem/Motivation

Error messages are crucial because they provide specific information about what went wrong, helping you quickly identify and resolve issues. Detailed error messages can pinpoint the exact part of the configuration or code that caused the problem. Error messages can include links or references to documentation, guiding you on how to fix the issue promptly.

While testing the Ratify policy for signature verification, we noticed that some error messages were difficult to comprehend from the user perspective.

Error message example 1:

```text
time=2024-07-17T16:28:16.939576441Z level=warning msg=Original Error: (Original Error: (HEAD "https://roacr.azurecr.io/v2/net-monitor/manifests/v2": GET "https://roacr.azurecr.io/oauth2/token?scope=repository%3Anet-monitor%3Apull&service=roacr.azurecr.io": response status code 401: unauthorized: authentication required, visit https://aka.ms/acr/authorization for more information.), Error: repository operation failure, Code: REPOSITORY_OPERATION_FAILURE, Plugin Name: oras), Error: get subject descriptor failure, Code: GET_SUBJECT_DESCRIPTOR_FAILURE, Plugin Name: oras, Component Type: referrerStore, Detail: failed to resolve the subject descriptor component-type=referrerStore go.version=go1.21.10 namespace= trace-id=34b27888-5402-443e-9836-77124c840561
```

This should have been an error message, instead, a warning appeared, containing multiple errors. The original error correctly described the source of the problem and pointed to a document for resolution, but the overall message was poorly formatted and the additional nested errors were redundant.

Error message example 2:

```text
"verifierReports": [
  {
        subject: "docker.io/library/hello-world@sha256:1408fec50309afee38f3535383f5b09419e6dc0925bc69891e79d84cc4cdce6",
        "isSuccess": false,
        "message: "verification failed: Error: no verifier report, Code: NO_VERIFIER _REPORT, Component Type: executor, Description: No verifier report was generatec preventing access to the registry, or the absence of appropriate verifiers corresponding to the referenced image artifacts."
  }
]
```

The json object "verifierReports" appears in the info level logs, and its format is organized and helpful for policy engines dealing with error information as needed. Yet, the message field is like other error messages and lacks clarity on the problem, its cause, and remediation methods. We also observed that the `verifierReports` contains different supported properties when compared with the Ratify Config policy and Ratify Rego policy, which is inconsistent.

Error message example 3:

```text
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: huishwabbit1.azurecr.io/test8may24@sha256:c780036bc8a6f577910bf01151013aaa18e255057a1653c76d8f3572aa3f6ff6
```

The policy engine, for instance, Gatekeeper, has produced this error message using a constraint template supplied by Ratify. It is the responsibility of the policy owner to tailor the error message to their requirements, however, some parts of the message extracted from Ratify's verification reports. In this example, the error message is not clear to users regarding the meaning of 'Subject', and fails to specify the type of verification error, such as whether it relates to signature verificatioin or SBOM verification. Additionally, reasons behind the errors are not provided. Furthermore, users may not be able to locate these errors in the complete AKS logs because simply searching the artifact digest does not pinpoint the exact logs needed.

Further findings covering a range of cases such as KMP, Store, Verifier, Policy configuration, access control, and signature verification issues are recorded at https://hackmd.io/@H7a8_rG4SuaKwzu4NLT-9Q/rkMLwv1F0. This link will be refreshed when the document is transferred to the Ratify repository.

To summarize, three areas require enhancement:

- Error message were not well formatted, refer to example 1
- Error messages within `verifierReport` wee not well formatted, refer to example 2
- Ratify did not supply enough data for policy engine to produce error messages , see example 3

The document aims to provide solutions and guidelines to improve error messages in different scenarios.

## Scenarios

### Error messages that shown in the Ratify logs

Alice works as a DevOps engineer at Contoso. She set up tasks to deploy containerized apps onto Kubernetes clusters. The cluster is assigned with the policy to deny the deployment of images that don't pass policy evaluation including verification of signature, SBOM, vulnerability reports and other image metadata. When policy evaluation fails, Alice wants to see concise, clear and actionable error messages in K8s logs. The error messages contain error description, error reasons and mitigation solutions, allowing her to address issues promptly.

### Error messages shown in verifier reports that are consumed by the policy engine

Bob works as a software engineer in the Policy team at Contoso. Bob write policies that are used at admission in K8s clusters to evaluate images by utilizing verifier reports that are produced by the Ratify service. Bob expects a verifier report to be machine readable, so that he can extract the needed data from the report. If a verifier report indicates the verification fails, Bob expects the verifier report can provide a concise, clear and actionable error message which includes error description, reasons and mitigation solutions. The policy will utilize these error messages to return final error messages to users or tasks that trigger the image deployment, for example using helm commands.

### Error messages returned by Ratify CLI commands

Gina is a software engineer on the CI/CD team at Contoso, where she creates pipeline tasks incorporating Ratify CLI commands to assess images according to policies. Should a policy check not pass, the corresponding images are prevented from progressing in the pipeline. Gina relies on Ratify CLI to produce concise, clear and actionable error messages that clearly explain what the error is, why it happens, and the actions needed to resolve them.

## Proposed solutions
