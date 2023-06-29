# Store/Verifier Providers
The framework uses a provider model for extensibility to support different types of referrer stores and verifiers. It supports two types of providers. 

Built-In/Internal providers are available in the source of the framework and are registered during startup using the ```init``` function. Referrer store using [ORAS](https://github.com/oras-project/oras) and signature verification using [notaryv2](https://github.com/notaryproject/notation) are currently available as the built-in providers in the framework and are managed by the framework.

External/Plugin providers are external to the source and process of the framework. They are registered as binaries that implement the plugin specification for the corresponding provider. The framework will locate these binaries in the configured paths and invoke them by passing the required parameters as per the specification. [SBOM](..\plugins\verifier\sbom) verification and signature verification using [cosign](..\plugins\verifier\cosign) libraries are examples of the providers that are implemented as plugins.

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
        "policy": "package ratify.policy\ndefault valid := false\nvalid {\n  not failed_verify(input)\n}\nfailed_verify(reports) {\n  [path, value] := walk(reports)\n  value == false\n  path[count(path) - 1] == \"isSuccess\"\n}"
    }
},
```
- The `name` field is REQUIRED and MUST match the name of the registered policy provider.
- One of `policyPath` and `policy` fields MUST be specified.
    - `policyPath`: path to the Rego policy file. The file MUST be a valid Rego policy file.
    - `policy`: Rego policy as a string. The string MUST be a valid Rego policy.

### Rego Policy Usage
Ratify embeds OPA engine inside the executor to provide a built-in policy provider. This feature is behind the `RATIFY_USE_REGO_POLICY` feature flag. And if users want to offload policy decison-making to Gatekeeper, they can enable `RATIFY_PASSTHROUGH_MODE` which will bypass OPA engine embedded in Ratify. In this mode, Ratify will NOT make the decision but only pass the verification results to Gatekeeper for making the decision. Note that verification results returned while switching Rego policy/config policy are a bit different.

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

To enable Rego policy provider for a Ratify server, the value of `RATIFY_USE_REGO_POLICY` MUST be set to true in the helm chart.

If Ratify is used as command line, `RATIFY_USE_REGO_POLICY=1` MUST be passed to the corresponding command1.

### Rego Policy Requirements

There are some special requirements on Rego policy used by Ratify. The package declaration MUST be `package ratify.policy`, and there MUST be a boolean variable `valid` declared. The variable `valid` MUST be set to `true` if the overall verification is successful, and `false` otherwise. The `input` is a Json object that contains the verification results of all reference artifacts. An example of the `input` is shown below:
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
              "verifierName": "notaryv2",
              "verifierType": "verifier-notary",
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
An example of a Rego policy is shown below:
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
### Examples:

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

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).