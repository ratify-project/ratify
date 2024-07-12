# Periodic Key & Certificate Retrieval for KMP

Author: Josh Duffney (@duffney)

Tracked issues in scope:

- https://github.com/ratify-project/ratify/issues/1131

Proposal ref:

- https://github.com/ratify-project/ratify/blob/dev/docs/proposals/Automated-Certificate-and-Key-Updates.md

## Problem Statement

In v1.2.0 and earlier, Ratify does not support automatic refreshing, requiring manual updates to the KMP resource to accommodate key and certificate changes. This also means that Ratify continues to use the cached versions of keys or certificates, leading to potential issues such as:

**Signature Verification Failures**: When images are signed with a newly rotated key version, Ratify’s cached key version fails to verify these signatures.

**Persisting Old Images**: Signature verification may fail for older images if Ratify only caches the latest key version.

**Usage of Disabled Keys**: Disabled keys, perhaps due to compromise, may still be used by Ratify from its cache, posing security risks.

Manual updates to KMP resources to accommodate key rotation are cumbersome and prone to misconfigurations, potentially causing image verification failures and service downtime.

## Proposed Solution

To address these challenges, this proposal suggests automating the update process of KMP resources in Ratify. This can be achieved by implementing a requeue mechanism on the KMP resource at a user defined interval.

## Implementation Details

Kubernetes has a built-in feature to requeue the controller's reconcile methods, which is responsible for populating the certificate and key values from the providers in Ratify's KMP resources. This feature is used by passing `{Requeue: true, RequeueAfter: interval}` to the `ctrl.Request` return by the KMP controller's reconcile method.

However, since not all of the providers support being refreshed. Inline, for example, would not benefit from being requeued after the creation of the resource. And for that reason, two new fields will be added to the spec of the CRD, interval and refreshable. Interval defined the value to wait inbetween refreshes and refreshable indicates if the resources can be refreshed. These values determine the value passed to the `ctrl.Result` returned by the reconcile method of the resource.

Refreshing the resources consists of calling the `getCertificate` and `getKeys` methods of the provider configured per resource. Meaning, if a resource is not refreshable these methods will only be called once to set up the resource and populate the in-memory maps that contain the keys and certificates configured on the provider. And if the provider is refreshable the get methods will be called again each time the interval is triggered.

An example of this implementation can be found below:

```go
// keymanagementprovider_controller.go
func (r *KeyManagementProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    //Do not requeue
    return ctrl.Result{}
    //Requeue
    return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 10}, nil
}
```

```yml
## Example KMP Resource
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementProvider
metadata:
name: keymanagementprovider-akv
spec:
  type: azurekeyvault
  interval: 1 #defined in minutes
  refreshable: true #indicates that the provider is able to refresh state
  parameters:
    vaultURI: https://${AKV_NAME}.vault.azure.net/
    keys:
      - name: ${KEY_NAME}
    tenantID: ${TENANT_ID}
    clientID: ${IDENTITY_CLIENT_ID}
---
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementProvider
metadata:
name: keymanagementprovider-inline
spec:
  type: inline
  #   refreshable: false (default value is false and do not need to be explicity stated)
  parameters:
    contentType: key
    value: ${Public_Key}
```

## Dev Work Items

Suggested steps to implement the proposed solution:

- Add `interval` and `refreshable` fields to the spec of the KMP CRD
- Implement a `refresh` interface that encapsulates the reconcile logic

## Open Questions

- How frequently should Ratify fetch the latest key or certificate versions from the KMS?
- How should Ratify support manual updates to the KMP cache?

## Future Considerations

- Event-Driven Key & Certificate Retrieval: Implementing an event-driven mechanism to fetch the latest key or certificate versions from the KMS based on specific triggers or events.
