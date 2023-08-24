# Store/Verifier Providers
The framework uses a provider model for extensibility to support different types of referrer stores and verifiers. It supports two types of providers. 

Built-In/Internal providers are available in the source of the framework and are registered during startup using the ```init``` function. Referrer store using [ORAS](https://github.com/oras-project/oras) and signature verification using [notation](https://github.com/notaryproject/notation) are currently available as the built-in providers in the framework and are managed by the framework.

External/Plugin providers are external to the source and process of the framework. They are registered as binaries that implement the plugin specification for the corresponding provider. The framework will locate these binaries in the configured paths and invoke them by passing the required parameters as per the specification. [SBOM](../../plugins/verifier/sbom) verification and signature verification using [cosign](../../plugins/verifier/cosign/) libraries are examples of the providers that are implemented as plugins.

# Policy Providers

Ratify implements an extensible policy provider interface allowing for different policy providers to be created and registered. The policy provider to be used is determined by the policy plugin specified in the `policy` section of the configuration. 

Currently, Ratify supports a Configuration based Policy Provider named `configPolicy`.

## How is the Policy Provider used in Ratify execution?

The executor is the "glue" that links all Ratify plugin-based components such as the verifiers, referrer stores, and policy providers. The policy provider is primarily invoked for each reference artifact discovered AFTER a configured verifier is found that can be used to verify that particular artifacy type. If a reference artifact is found that does not have a corresponding verifier configured that can perform verification, the executor will ignore this artifact. As a result, this unverifiable artifact will NOT influence the final overall verification success determined by the policy provider.

## Rego Policy Provider

The Rego Policy Provider is a built-in policy provider that uses [Open Policy Agent](https://www.openpolicyagent.org/) (OPA) and [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) to implement a policy provider. The Rego Policy Provider is a plugin that is registered with Ratify and is invoked by the executor to determine the overall verification success if required. The Rego Policy Provider is configured using the `policy` section of the configuration.

```json
"policy": {
    "version": "1.0.0",
    "plugin": {
        "name": "regoPolicy",
        "policyPath": "",
        "policy": "package ratify.policy\ndefault valid := false\nvalid {\n  not failed_verify(input)\n}\nfailed_verify(reports) {\n  [path, value] := walk(reports)\n  value == false\n  path[count(path) - 1] == \"isSuccess\"\n}",
        "passthroughEnabled": false
    }
},
```
- The `name` field is REQUIRED and MUST match the name of the registered policy provider.
- One of `policyPath` and `policy` fields MUST be specified.
    - `policyPath`: path to the Rego policy file. The file MUST be a valid Rego policy file.
    - `policy`: Rego policy as a string. The string MUST be a valid Rego policy.
- The `passthroughEnabled` field is optional, which defaults to be `false` if not provided, in thise case, Ratify will return the verification results with the decision to Gatekeeper. If `passthroughEnabled` is set to `true`, the executor will NOT make the decision but only pass the verification results to Gatekeeper for making the decision.

### Rego Policy Usage
Ratify embeds OPA engine inside the executor to provide a built-in policy provider. There are 2 approaches to enable this feature as an add-on service.

1. Set the helm chart value of `policy.useRego` to `true` while deploying Ratify.
2. Apply a Policy Custom Resource with Rego Policy if the service is up. e.g.
```bash
kubectl apply -f ./config/samples/policy/config_v1alpha1_policy_rego.yaml
```

And if Ratify is used as command line tool, users MUST provide a config with Rego Policy. Check `test/bats/tests/config/config_rego_policy_notation_leaf_cert.json` as an example.

Note that verification results returned are different while switching Rego policy/config policy. 

When Rego Policy is selected, the Verification Response follows `1.0.0` version. [1.0.0](../reference/verification-result-version.md#1.0.0) provides definition and example usage of the response.

### Rego Policy Requirements

There are some special requirements on Rego policy used by Ratify. The package declaration MUST be `package ratify.policy`, and there MUST be a boolean variable `valid` declared. The variable `valid` MUST be set to `true` if the overall verification is successful, and `false` otherwise. The `input` is a Json object that contains the verification results of all reference artifacts.

### Rego Policy Examples
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

#### Example 1
Require all reference artifacts associated with subject image to be verify successfully.
Check equivalent [Config Policy](#config-policy-examples)

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

#### Example 2
Require at least one reference artifact of the same type to verify succesfully. (relaxes the default policy to 'any').
Check equivalent [Config Policy](#config-policy-examples)

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

## Config Policy Provider

```
...
"policy": {
    "version": "1.0.0",
    "plugin": {
        "name": "configPolicy",
        "artifactVerificationPolicies": {
            "application/vnd.cncf.notary.signature": "any"
        }
    }
},
...
```

- The `name` field is REQUIRED and MUST match the name of the registered policy provider
- `artifactVerificationPolicies`: map of artifact type to policy; each entry in the map's policy must be satisfied for Ratify to return true.
    - `any`: policy that REQUIRES at least one artifact of specified type to verify to `true` 
    - `all`: policy that REQUIRES all artifacts of specified type to verify to `true``
- Default policy:
    - The `default` policy applies to unspecified artifact types. The `default` policy is set to `all`. Thus, all unspecified artifact types must have all successful verification results for an overall success result.
    - The `default` policy can be overridden to `any` in the map:
        
        ```
        ...
        "policy": {
            "version": "1.0.0",
            "plugin": {
                "name": "configPolicy",
                "artifactVerificationPolicies": {
                    "default": "any"
                }
            }
        },
        ...
        ```
### Config Policy Examples:

- Require all reference artifacts associated with subject image to be verify successfully:
    ```
    ...
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy"
        }
    },
    ...
    ```
- Require at least one reference artifact of the same type to verify succesfully. (relaxes the default policy to 'any'):
    ```
    ...
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "default": "any"
            }
        }
    },
    ...
    ```
- For a specific artifact type, relax requirement so only one success is needed for artifacts of that type:
    ```
    ...
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.cncf.notary.signature": "any"
            }
        }
    },
    ...
    ```

The Verification response follows `0.1.0` version. Check definition and examples in [0.1.0](../reference/verification-result-version.md#0.1.0).
## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).