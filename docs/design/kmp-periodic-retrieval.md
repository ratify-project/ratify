# Periodic Key & Certificate Retrieval for KMP
Author: Josh Duffney (@duffney)

Tracked issues in scope:
- https://github.com/ratify-project/ratify/issues/1131

Proposal ref:
- https://github.com/ratify-project/ratify/blob/dev/docs/proposals/Automated-Certificate-and-Key-Updates.md

## Problem Statement

Ensuring the integrity and authenticity of images through signing and verification involves using tools like Notation with code-signing certificates in Azure Key Vault (AKV) or Cosign with key pairs stored in AKV. In Kubernetes (K8s), verification of these signed images with Ratify requires configuring corresponding certificates or keys using a custom resource called KeyManagementProvider (KMP). However, a significant challenge arises due to the key and certificate rotation practices within a Key Management System (KMS).

In v1.2.0 or earlier, Ratify does not support automatic key rotation, requiring manual updates to the KMP resource to accommodate key rotation. Meaning that Ratify continues to use the cached versions of keys or certificates, leading to potential issues such as: 

**Signature Verification Failures**: When images are signed with a newly rotated key version, Ratify’s cached key version fails to verify these signatures.

**Persisting Old Images**: Signature verification may fail for older images if Ratify only caches the latest key version.

**Usage of Disabled Keys**: Disabled keys, perhaps due to compromise, may still be used by Ratify from its cache, posing security risks.

Manual updates to KMP resources to accommodate key rotation are cumbersome and prone to misconfigurations, potentially causing image verification failures and service downtime.

## Proposed Solution

To address these challenges, this proposal suggests automating the update process of KMP resources in Ratify. This can be achieved by implementing a mechanism within Ratify to dynamically fetch and update the latest key or certificate versions upon rotation without requiring manual intervention. 

## Implementation Details

### Periodic Key & Certificate Retrieval

To automate the update process of KMP resources, Ratify can periodically fetch the latest key or certificate versions from the KMS. This can be achieved by adding an interval field to the operator configuration and implementing a `reconciliation loop` within the KMP controller to periodically fetch the latest key or certificate versions from the KMS.

An example of this implementation can be found below:

```go
// keymanagementprovider_controller.go
func (r *KeyManagementProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    //.....
    return ctrl.Result{RequeueAfter: getRequeueDuration(now, updateInterval)}, nil
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
parameters:
    vaultURI: https://${AKV_NAME}.vault.azure.net/
    keys:
    - name: ${KEY_NAME}
    tenantID: ${TENANT_ID}
    clientID: ${IDENTITY_CLIENT_ID}
    updateInterval: 1h
```

### Key & Certificate Rotation

**Scenario 1**: Versions Not Specified for Keys or Certificates

```yml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementProvider
metadata:
name: keymanagementprovider-akv
spec:
type: azurekeyvault
parameters:
    vaultURI: https://${AKV_NAME}.vault.azure.net/
    keys:
    - name: ${KEY_NAME}
    tenantID: ${TENANT_ID}
    clientID: ${IDENTITY_CLIENT_ID}
```

1. The Reconiliation Loop retrieves the most recent key or certificate version from the provider and verifies the status of existing versions to ensure they are enabled and not disabled or revoked.
2. The KMP provider retrieves the most recent key or certificate version.
3. If the most recent version differs from the cached one, the cache is updated with the latest version of the key or certificate.
4. If the KMP cache contains more than three versions of the key or certificate, the oldest version is deleted.

When a key or certificate version isn't provided in the KMP resource, Ratify should fetch the latest key or certificate version from the KMS. By default, the KPM will store the last three versions of the key or certificate.

**Scenario 2**: Versioned Keys or Certificates

When a key or certificate version is provided in the KMP resource, Ratify should fetch the specified key or certificate version from the KMS.

```yml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementProvider
metadata:
name: keymanagementprovider-akv
spec:
type: azurekeyvault
parameters:
    vaultURI: https://${AKV_NAME}.vault.azure.net/
    keys:
    - name: ${KEY_NAME}
        version: ${KEY_VER1}
    tenantID: ${TENANT_ID}
    clientID: ${IDENTITY_CLIENT_ID}
```

1. When the Reconiliation Loop is activated, it prompts the KMP provider to retrieve the designated key or certificate version.
2. The KMP provider then confirms the validity of the designated key or certificate version, ensuring it is enabled and not disabled or revoked.
3. In case the specified latest version differs from the cached one, info logs are generated, and the system continues to utilize the cached version specified by the config.
4. If the specified version matches the cached one, no further action is required.
5. If the specified version cannot be located, the key or certificate is removed from the cache.

### Key Disabling or Certificate Revocation

When a key or certificate is disabled or revoked in the KMS, Ratify should remove the key or certificate from the cache and generate warning logs. This ensures that Ratify does not use disabled keys or revoked certificates, thereby enhancing security.

1. The Reconiliation Loop retrieves the most recent key or certificate version from the provider and verifies the status of existing versions to ensure they are enabled and not disabled or expired.
2. If the KMP provider detects that a key or certificate has been disabled or revoked, it removes the key or certificate from the cache and generates warning logs.

## Update certificates and keys manually

In some cases, it may be necessary to manually update the certificates or keys in the KMP cache. This can be achieved by updating the KMP resource with the new key or certificate version.

TODO: 

## Dev Work Items

Suggested steps to implement the proposed solution:
- Implement the `reconciliation loop` within the KMP controller to periodically fetch the latest key or certificate versions from the KMS.
- Update KMP providers to retrieve the most recent key or certificate version and verify the status of existing versions.
- Implement logic in verifiers to utilize multiple versions of keys or certificates.
- Support manually triggering the reconcile method. 

## Open Questions

- How frequently should Ratify fetch the latest key or certificate versions from the KMS?
- Should the reconciliation interval be defined at the operator level or the KMP resource level?
- Do verifiers support the use of multiple versions of keys or certificates?
- How should Ratify support manual updates to the KMP cache?

## Future Considerations

- Event-Driven Key & Certificate Retrieval: Implementing an event-driven mechanism to fetch the latest key or certificate versions from the KMS based on specific triggers or events.
