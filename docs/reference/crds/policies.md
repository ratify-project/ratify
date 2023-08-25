A `Policy` resource defines a policy evaluating the verification results for a subject.

View more CRD samples [here](../../../config/samples/policy). The `metadata.name` is the policy name which would be applied. Common properties:
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Policy
metadata:
  name: "configpolicy"
spec:
  parameters: required. Parameters specific to this policy
```
There would be exactly one CR existing in the cluster. That means if users apply a new CR, the old one would be replaced.

Note: `metadata.name` MUST be `configpolicy` or `regopolicy` per the usage.

## configpolicy
Sample spec:
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Policy
metadata:
  name: "configpolicy"
spec:
  parameters:
    artifactVerificationPolicies:
      "application/vnd.cncf.notary.signature": "any"
      default: "any"
```
| Name | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- |
| artifactVerificationPolicies | yes | Map of artifact type to policy; each entry in the map's policy must be satisfied for Ratify to return true | "" |
| default | no | The `default` policy applies to unspecified artifact types. | "all" |
| `application/vnd.cncf.notary.signature` | no | It could be any artifact type that is supported by Ratify. | There is no default value, users must specify `any` or `all` |

## regopolicy
Sample spec:
```yml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Policy
metadata:
  name: "regopolicy"
spec:
  parameters:
    passthroughEnabled: false
    policy: |
      package ratify.policy

      default valid := false

      # all artifacts MUST be valid
      valid {
        not failed_verify(input)
      }

      # all reports MUST pass the verification
      failed_verify(reports) {
        [path, value] := walk(reports)
        value == false
        path[count(path) - 1] == "isSuccess"
      }

      # each artifact MUST have at least one report
      failed_verify(reports) {
        [path, value] := walk(reports)
        path[count(path) - 1] == "verifierReports"
        count(value) == 0
      }
```
| Name | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- |
| passthroughEnabled | no | If set to true, Ratify will NOT make the decision but pass verifier reports to Gatekeeper. | false |
| policy | no | The policy language that defines the policy. | "" |
| policyPath | no | The path to the policy file if policy is mounted as a volume | "" |

Note: Users MUST provide at least one of `policy` and `policyPath`. If both are specified, `policy` will be used. 