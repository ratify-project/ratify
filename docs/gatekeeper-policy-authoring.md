# Gatekeeper Policy Authoring
Ratify can be deployed using passthrough execution mode behind admission
controllers such as Gatekeeper. When deployed in this manner Ratify will provide
the results of all verifiers back to Gatekeeper so that policies can be authored
in rego.

Rego policies utilizing external data are authored in exactly the same way as
traditional rego policies, but need to know the structure of the response from
the external data provider in order to take advantage of it.

## Rego References
These are some helpful references for working with rego:
- [Rego Policy Language](https://www.openpolicyagent.org/docs/latest/policy-language/)
- [Rego Policy Reference](https://www.openpolicyagent.org/docs/latest/policy-reference/)
- [Rego Built-in Functions](https://www.openpolicyagent.org/docs/latest/policy-reference/#built-in-functions)
- [Gatekeeper External Data](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata)
- [Gatekeeper Policy Debugging](https://open-policy-agent.github.io/gatekeeper/website/docs/debug)
- [Rego Playground](https://play.openpolicyagent.org/)

The rego playground is a great way to test out policies against responses from
Ratify. Use the steps from the Gatekeeper debugging reference to print the
response from Ratify. This can then be used as the input for rego playground. In
rego playground you can replace the external_data call with the input:
```rego
package ratify
# You would use the below when running the rego using external_data
remote_data := response {
  images := [img | img = input.review.object.spec.containers[_].image]
  response := external_data({"provider": "ratify-provider", "keys": images})
}
# Use this to mimic the behavior of external_data within rego playground
remote_data := response {
  response := input
}
```

## Ratify Response in Rego
Let's assume we have a registry with the following image and associated
artifacts:
```
localhost:5000/net-monitor:v1
    - sbom
    - notaryv2 signature
    - cosign signature
```

We will try to deploy the following pod spec:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
    - name: net-monitor
      image: localhost:5000/net-monitor:v1
    - name: missing-image
      image: localhost:5000/missing-image:v1
```
Note that one of the images listed in the pod spec does not actually exist in
the registry. This is to give an example of something that might generate an
error message in one of the response fields.

Given the following rego snippet:
```rego
package ratify
# Get data from Ratify
remote_data := response {
  images := [img | img = input.review.object.spec.containers[_].image]
  response := external_data({"provider": "ratify-provider", "keys": images})
}
```

The `remote_data` document contents would look something like this (NOTE: this
is a sample response to explain the structure of the response data and the
actual values of fields may differ):
```json
{
  "errors": [
    ["localhost:5000/missing-image:v1", "The subject did not exist"]
  ],
  "responses": [
    [
      "localhost:5000/net-monitor:v1",
      {
        "isSuccess": true,
        "verifierReports": [
          {
            "artifactType": "application/x.example.sbom.v0",
            "isSuccess": true,
            "message": "SBOM verification success. The contents is good.",
            "name": "sbom",
            "subject": "localhost:5000/net-monitor:v1"
          },
          {
            "artifactType": "application/vnd.cncf.notary.v2.signature",
            "extensions": {
              "Issuer": "CN=localhost:5000,O=Ratify,L=Seattle,ST=Washington,C=US",
              "SN": "CN=localhost:5000,O=Ratify,L=Seattle,ST=Washington,C=US"
            },
            "isSuccess": true,
            "message": "signature verification success",
            "name": "notaryv2",
            "subject": "localhost:5000/net-monitor:v1"
          },
          {
            "artifactType": "org.sigstore.cosign.v1",
            "isSuccess": true,
            "message": "cosign verification success. valid signatures found",
            "name": "cosign",
            "subject": "localhost:5000/net-monitor:v1"
          }
        ]
      }
    ]
  ],
  "status_code": 200,
  "system_error": ""
}
```
Most of the response structure is dictated by the Gatekeeper external data
[ProviderResponse](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata#api-v1alpha1)
interface. This structure includes:
- `errors`: An array of 0-N elements where N is typically the number of subjects
  from the external data request. This array contains nested arrays where each
  nested array contains the subject as the first element and an error message as
  the second element.
- `responses`: An array of 0-N elements where N is typically the number of
  subjects from the external data request. This array contains nested arrays
  where each nested array contains the subject as the first element and a
  value as the second element. The type of the value is dictated by the external
  data provider.
- `status_code`: The HTTP status code returned from the request to the external
  data provider.
- `system_error`: A system level error that can either be returned by the
  external data provider or result from some communication issue with the
  external data provider.

When Ratify is configured in passthrough execution mode an array of each
`VerifyResult` for a subject is included in the `response` field and the overall
`isSuccess` value will always be true. This allows the policy to be entirely
authored in the rego using all the results from the individual verifiers.

Each verifier produces a `VerifierResult` with the following potential fields:
```go
type VerifierResult struct {
	Subject       string           `json:"subject,omitempty"`
	IsSuccess     bool             `json:"isSuccess"`
	Name          string           `json:"name,omitempty"`
	Message       string           `json:"message,omitempty"`
	Extensions    interface{}      `json:"extensions,omitempty"`
	NestedResults []VerifierResult `json:"nestedResults,omitempty"`
	ArtifactType  string           `json:"artifactType,omitempty"`
}
```
The most commonly used fields are:
- `Name`: The name of the verifier
- `IsSuccess`: The result of the verifier
- `Message`: A message about the success or failure of the verifier
- `Extensions`: Additional metadata added to the result by the verifier

The `Extensions` field allows the verifiers to include additional metadata as
part of the response to use for policy evaluation (for example, the notaryv2
verifier includes the Issuer and SubjectName from the certificate as an extension field).
Currently, the best way to determine the extension fields for a verifier is to
check the source for that verifier.

## Sample Rego Policy
Let's look at a Ratify policy from the library in more detail. Below is the
rego from the `notaryv2issuervalidation` policy:
```rego
package notaryv2issuervalidation

# Get data from Ratify
remote_data := response {
  images := [img | img = input.review.object.spec.containers[_].image]
  response := external_data({"provider": "ratify-provider", "keys": images})
}

# Base Gatekeeper violation
violation[{"msg": msg}] {
  general_violation[{"result": msg}]
}

# Check if there are any system errors
general_violation[{"result": result}] {
  err := remote_data.system_error
  err != ""
  result := sprintf("System error calling external data provider: %s", [err])
}

# Check if there are errors for any of the images
general_violation[{"result": result}] {
  count(remote_data.errors) > 0
  result := sprintf("Error validating one or more images: %s", remote_data.errors)
}

# Check if the success criteria is true
general_violation[{"result": result}] {
  subject_validation := remote_data.responses[_]
  subject_validation[1].isSuccess == false
  result := sprintf("Subject failed verification: %s", [subject_validation[0]])
}

# Check that signature result for Issuer exists
general_violation[{"result": result}] {
  subject_results := remote_data.responses[_]
  subject_result := subject_results[1]
  notaryv2_results := [res | subject_result.verifierReports[i].name == "notaryv2"; res := subject_result.verifierReports[i]]
  issuer_results := [res | notaryv2_results[i].extensions.Issuer == input.parameters.issuer; res := notaryv2_results[i]]
  count(issuer_results) == 0
  result := sprintf("Subject %s has no signatures for certificate with Issuer: %s", [subject_results[0], input.parameters.issuer])
}

# Check for valid signature
general_violation[{"result": result}] {
  subject_results := remote_data.responses[_]
  subject_result := subject_results[1]
  notaryv2_results := [res | subject_result.verifierReports[i].name == "notaryv2"; res := subject_result.verifierReports[i]]
  notaryv2_result := notaryv2_results[_]
  notaryv2_result.isSuccess == false
  notaryv2_result.extensions.Issuer == input.parameters.issuer
  result = sprintf("Subject %s failed signature validation: %s", [subject_results[0], notaryv2_result.message])
}
```

There are a handful of rules defined here that serve different purposes.

The first rule creates a document called `remote_data` by making a call to the
Ratify external data provider. This document is then utilized in other rules to
evaluate the results.

The `violation` rule is a Gatekeeper construct and is used by Gatekeeper to
determine rule violations. If a message is not present after evaluating the rego
the admission request will pass. From this rule we call all the different types
of `general_violations` that we have defined.

Finally, there are a collection of `general_violation` rules. Since these rules
all share the same name and signature they are iteratively building a set of
messages that will be bubbled up to the main `violation` rule. Each of them is
checking for different criteria that would indicate a violation.

The first `general_violation` is satisfied when there is a `system_error` field
under the `remote_data` document and when the field is not an empty string. When
the body of the rule is satisfied it will generate a result defined by the
`sprintf` statement. This result is then appended to the `msg` in the
`violation` rule and will be returned to the user.

Each subsequent `general_violation` is checking for different criteria that
would indicate a violation using standard rego.

When building custom policies the top four rules can be reused as they are used
to handle data fetching and general error cases. Custom `general_violation`
rules can then be added to handle specific violations.
