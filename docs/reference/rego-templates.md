# Rego Templates

Rego Policy is much more powerful than Config Poliy, but it's also more complex. To make it easier to write Rego Policy, we provide a few templates for common use cases.

## Rego Policy Requirements

There are some special requirements on Rego policy used by Ratify. The package declaration MUST be `package ratify.policy`, and there MUST be a boolean variable `valid` declared. The variable `valid` MUST be set to `true` if the overall verification is successful, and `false` otherwise. The `input` is a Json object that contains the verification results of all reference artifacts.

## Examples

Here is a sample of the `input`:
```json
{
  "verifierReports": [
    {
      "artifactType": "org.example.sbom.v0",
      "subject": "test.azurecr.io/test/hello-world:v1",
      "referenceDigest": "test.azurecr.io/test/hello-world@sha256:22222",
      "verifierReports": [
        {
          "verifierName": "sbom",
          "verifierType": "sbom-verifier",
          "isSuccess": false,
          "message": "error msg",
          "extensions": {}
        },
        {
          "verifierName": "schemavalidator",
          "verifierType": "verifier-schemavalidator",
          "isSuccess": true,
          "message": "",
          "extensions": {}
        }
      ],
      "nestedReports": [
        {
          "artifactType": "application/vnd.cncf.notary.signature",
          "subject": "test.azurecr.io/test/hello-world@sha256:123",
          "referenceDigest": "test.azurecr.io/test/hello-world@sha256:33333",
          "verifierReports": [
            {
              "verifierName": "notation",
              "verifierType": "verifier-notation",
              "isSuccess": true,
              "message": "",
              "extensions": {
                "Issuer": "testIssuer",
                "SN": "testSN"
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

### Example 1
Require all reference artifacts associated with subject image to be verify successfully.
Check equivalent [Config Policy](./providers.md#config-policy-examples)

```rego
package ratify.policy

default valid := false

valid {
  not failed_verify(input)
}

failed_verify(reports) {
  [path, value] := walk(reports)
  value == false
  path[count(path) - 1] == "isSuccess"
}

failed_verify(reports) {
  [path, value] := walk(reports)
  path[count(path) - 1] == "verifierReports"
  count(value) == 0
}
```

### Example 2
Require at least one reference artifact of the same type to verify succesfully. (relaxes the default policy to 'any').
Check equivalent [Config Policy](./providers.md#config-policy-examples)

```rego
package ratify.policy

default valid := false

valid if {
	not failed_verify(input)
}

failed_verify(reports) if {
	newReports := {"nestedReports": reports.verifierReports}
	has_subject_failed_verify(newReports)
}

has_subject_failed_verify(nestedReports) if {
	[path, value] := walk(nestedReports)
	not artifact_type_pass_verify(value)
	path[count(path) - 1] == "nestedReports"
}

# at least one artifact of the same type passed verification
artifact_type_pass_verify(nestedReports) if {
	count_artifact_type(nestedReports) == count_successful_artifact_type(nestedReports)
}

count_artifact_type(nestedReports) := number if {
	artifact_types := {x |
		some i
		x := nestedReports[i].artifactType
	}

	number := count(artifact_types)
}

count_successful_artifact_type(nestedReports) := number if {
	artifact_types := {x |
		some i
		x := nestedReports[i].artifactType
		artifact_pass_verify(nestedReports[i].verifierReports)
	}

	number := count(artifact_types)
}

# an artifact has at least one successful report
artifact_pass_verify(verifierReports) if {
	verifierReports[_].isSuccess == true
}

# should be at least one verifier report that is successful
failed_verify(reports) if {
	[path, value] := walk(reports)
	path[count(path) - 1] == "verifierReports"
	count(value) == 0
}
```

### Example 3
I trust multiple identities, and need to find at least one valid signature signed by any identity.

Prerequisites:

1. The artifact is only signed by Notation, there might be a few signatures attached to the artifact.
2. Users has configured different Notation verfiers, and each verifier has a different identity.


```rego
package ratify.policy

import future.keywords.in

default valid := false

valid {
  not failed_verify(input)
}

# Fail the verification if no signatures are signed by trusted identity.
failed_verify(reports) {
  not has_trusted_signature(reports.verifierReports)
}

# Fail the verification if no signatures are present.
failed_verify(reports) {
  [path, value] := walk(reports)
  path[count(path) - 1] == "verifierReports"
  count(value) == 0
}

# There is a signature signed by trusted identity.
has_trusted_signature(reports) {
	some report in reports
  report.artifactType == "application/vnd.cncf.notary.signature"
  pass_validation(report.verifierReports)
}

pass_validation(reports) {
	some report in reports
  report.isSuccess == true
}
```

### Example 4
I trust multiple identities, and need to find at least one valid signature signed by each identity.

Prerequisites:

1. The artifact is only signed by Notation, there might be a few signatures attached to the artifact.
2. Users has configured different Notation verfiers, and each verifier has a different identity.

Note: The below Rego query currently does not work as expected as it's blocked by a known [bug](https://github.com/deislabs/ratify/issues/1066) which reports the same name for different Notation verifiers.

```rego
package ratify.policy

import future.keywords.in

default valid := false

valid {
  not failed_verify(input)
}

# Every verifier has verified one signature.
has_all_trusted_identities(reports) {
	# Get names of verifiers that pass the verification
  names := { name |
    reports[i].artifactType == "application/vnd.cncf.notary.signature"
    name := reports[i].verifierReports[j].name
      reports[i].verifierReports[j].isSuccess == true
  }
  # Get number of different verifiers
  verifier_number := [ number |
    some i
    reports[i].artifactType == "application/vnd.cncf.notary.signature"
    number := count(reports[i].verifierReports)
  ]
  
  verifier_number[0] == count(names)
}

# Fail the verification if one trust identity didn't sign any signatures.
failed_verify(reports) {
  not has_all_trusted_identities(reports.verifierReports)
}

# Fail the verification if no signatures are present.
failed_verify(reports) {
  [path, value] := walk(reports)
  path[count(path) - 1] == "verifierReports"
  count(value) == 0
}
```