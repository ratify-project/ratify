# Automated Certificate and Key Updates

## Problem/Motivation

When ensuring the integrity and authenticity of images, you can sign images using Notation with code-signing certificates in Azure Key Vault (AKV) or Cosign with key pairs stored in AKV. To verify these signed images with Ratify in K8s, users typically configure the corresponding certificates or keys using a custom resource called `KeyManagementProvider` (referred to as `KMP` for short). This configuration allows Ratify to retrieve the correct certificates or keys for verification.

In most cases, certificates and keys are securely managed within a Key Management System (KMS), such as AWS KMS, Azure Key Vault (AKV), HashiCorp Vault, or GCP KMS. As a proactive security best practice, both certificates and keys are regularly rotated within a KMS. This rotation can occur automatically or manually (On-demand) - For example, when certificates or keys are compromised. For most of KMS, after rotation, the cryptographic material associated with the key or certificate is updated while maintaining the same references (i.e., the key name or certificate name). This ensures compatibility with existing applications and services that rely on that certificate or key. For instance, if a certificate is rotated in AKV, the certificate name remains unchanged, but a new version of the certificate is created. The cryptographic key identifier in the new version differs from the previous version. Applications can continue referencing the certificate using the same name without breaking compatibility. 

In the current Ratify version (v1.2.0 or earlier), AKV is the only supported KMS provider. Consider AKV key configuration as an example: users can specify either the key name or a specific version of the key. If only the key name is configured, Ratify fetches the latest version of the configured key and caches it. When a specific version is configured, Ratify retrieves that precise version and caches it. However, unless users update the existing `KMP` resource to use a different key or key version, or delete the existing `KMP` and reapply it, the cached key versions remain unchanged. Consequently, when key rotation occurs or the key’s operational status changes, Ratify continues to rely on the cached version. This can lead to several issues:

- **Signature Verification Failures**
  - When images are signed with the latest version of the key, signature verification fails because the cached key is not updated.
  - Images signed with previous versions may persist for an extended period, but Ratify only caches the latest version. Consequently, signature verification may fail for images signed using older versions.
- **Disabled Keys Should Not Be Used**
  - If a key is disabled (due to compromise, for example), Ratify can still use the cached version for image verification. This poses significant security risks, as disabled keys should not be employed in any cryptographic operations.

To address these challenges, users need to manually update or reapply the `KMP` resource to trigger Ratify to retrieve the latest/specific versions of keys or certificates. This step can be cumbersome, especially considering automated key rotation. Additionally, users must keep the previous versions of keys configured for some time, as not all images are signed with the latest/specific versions. In practice, keeping up with these changes manually can be challenging, and misconfigurations may lead to image verification failures and unnecessary service downtime. 

Certificate rotation in AKV follows a similar process to key rotation, as described earlier, but its impact is less significant. According to the Notary Project specification, only root CA certificates are required for trust stores. The root CA certificate typically has a long validity and is unlikely to change during certificate rotation unless users switch to a different CA for issuing code-signing certificates. As a result, the cached root CA certificate can be used for quite a long period normally.

To address these problems, this document begins by comparing various key management providers. It then outlines scenarios, proposal and user experiences.

## Key Management Providers comparison

This section compares various KMPs that can be integrated with Ratify for signature verification scenarios:
- AWS KMS: https://docs.aws.amazon.com/kms/
- Azure Key Vault: https://learn.microsoft.com/en-us/azure/key-vault/ (already supported by Ratify)
- HCP Vault: https://developer.hashicorp.com/vault/docs/what-is-vault
- GCP KMS: https://cloud.google.com/kms/docs/key-management-service
- AWS Signer: https://docs.aws.amazon.com/signer/latest/developerguide/Welcome.html
- Azure Trusted Signing: https://learn.microsoft.com/en-us/azure/trusted-signing/overview

Here’s a comparison of various KMSs used for creating and managing keys/certificates for signing and verification:

|                            | **AWS KMS** | **Azure Key Vault** | **HCP Vault** | **GCP KMS** |
|----------------------------|-------------|---------------------|---------------|-------------|
| **Key Operations**         | Enable, Disable, Delete, Versioning, Rotate | Enable, Disable, Delete, Versioning, Rotate | Enable, Disable, Delete, Versioning, Rotate | Enable, Disable, Delete, Versioning, Rotate |
| **Key Rotation**           | Manual for asymmetric keys | Automatic/Manual | Automatic/Manual | Manual for asymmetric keys |
| **Rotation Result**        | New version created, old version remains valid | New version created, old version remains valid | New version created, old version remains valid | New version created, old version remains valid |
| **Events**                 | CloudWatch Events for key state changes | Azure Event Grid for key state changes | Custom implementation (e.g., using Vault's audit logs) | Cloud Pub/Sub for key state changes |
| **Certificate management** | Not supported | Supported | Supported | Not supported |

Additionally, here’s a comparison of fully managed code-signing services related to Ratify signature verification scenarios. The difference from KMS in previous table is that fully managed code-signing services do not require users to create keys/certificates, instead they managed the key/certificates for users and normally these key/certificates have short validity. For fully managed code-signing, normally users set up a profile for signing purpose. The profile identifier is similar to the key identifier that used by KMS in previous table. Users can revoke the profile, so that it cannot be used for signing an verification. 

|                        | **AWS Signer**                                   | **Azure Trusted Signing**             |
|------------------------|--------------------------------------------------|---------------------------------------|
| **Configuration**      | Signing profile                                  | Certificate profile                   | 
| **Revocation**         | Signature revocation, signing profile revocation | Certificate profile revocation        |
| **Signature Type**     | Notary Project signature                         | Notary Project signature              |

Lastly, here’s `KMP` resource availability in Ratify v1.2.0:

|                                       | **AWS KMS** | **Azure KV** | **HCP Vault** | **GCP KMS** | **AWS Signer** | **Azure Trusted Signing** |
|---------------------------------------|-------------|---------------------|---------------|-------------|------------------|---------------------------|
| Notary Project signature verification | N/A   | `azurekeyvault KMP` | N/A | N/A | N/A | N/A |
| Cosign signature verification         | N/A    | `azurekeyvault KMP` | N/A | N/A| N/A | N/A |

If users have keys or certificates stored in a KMS, which does not have a corresponding `KMP` resource in Ratify, user can download the key or certificate from the KMS and configure a `inline KMP` for verification.

## Scenarios

### Certificate rotation

Alice, an application engineer at Contoso LLC, focuses on securing containerized applications. She sets up the build pipeline to sign container images using certificates stored in a KMS. To ensure that only trusted images are deployed in Kubernetes clusters, she deploys Gatekeeper and Ratify. Ratify is configured to validate signatures using certificates in a KMS. Alice applies the security best practice that the certificate expires after certain periods, and automatic rotation is enabled before expiry. Alice wants Ratify to sync multiple versions of the certificate in AKV including latest version and previous versions. This way, Ratify can successfully verify images signed by different versions without requiring manual configuration adjustments. Additionally, she sets up the pipeline to update the signing certificate to the latest version for signing container images. By doing so, she ensures that container images are consistently signed with up-to-date certificates. The automatic handling of certificate updates enhances security and reduces the risk of service disruptions.

### Certificate revocation

A malicious user compromised specific versions of certificate. Alice promptly reported this incident to the Certificate Authority (CA). The CA added the compromised versions to the Certificate Revocation List (CRL) and responded to Online Certificate Status Protocol (OCSP) queries with the revoked status. As Alice expects, Ratify denies the deployment of images that were signed with revoked versions. Alice rotated the certificate manually resulting a new version of the certificate with new cryptographic materials. Alice wants Ratify to sync multiple versions of a certificate including latest version and previous versions, This way, it ensures successful verification of images signed with the latest version without requiring manual configuration adjustments. Additionally, the build pipeline was triggered to build and sign all the images with the new version of the certificate. By following these actions, Alice maintains security while seamlessly transitioning to the new certificate version.

### Key rotation

Bob, an application engineer at Wabbit Network LLC, focuses on securing containerized applications. He sets up the build pipeline to sign container images using key pairs in a KMS. To ensure that only trusted images are deployed in Kubernetes clusters, he deploys Gatekeeper and Ratify. Ratify is configured to validate signatures using public keys in the KMS. As a security best practice, Bob configured automated key rotation for keys in the KMS, resulting a new version of the key created before key expiry. Bob wants Ratify to sync multiple versions of a key including latest version and previous versions, allowing successful verification of images signed by different key versions without any manual configuration adjustments. Additionally, Bob sets up the pipeline to update the signing key to the latest version for signing container images. Bob knows that images are not allowed to be signed using keys that have expired in the KMS, but verification using expired key is allowed for verifying images that were signed at the time the key was valid. By following these actions, Bob maintains security while seamlessly supporting verification with the new version of key.

### Key disabling

A malicious user compromised the private key used for signing images. Bob promptly rotated the key manually, obtaining a new version, and disabled the compromised version. Bob wants Ratify to sync multiple versions of a key including latest version and previous versions, and excluding any disabled versions without requiring manual configuration adjustments, as a result, images signed with the disabled version fail signature verification and images signed with new version can be verified successfully. Additionally, the build pipeline was triggered to build and sign all the images with the new version of key. By following these actions, Bob maintains security while seamlessly transitioning to the new version of key.

### Update certificates and keys manually

In some scenarios, such as when the private key is compromised, both Alice and Bob prefer Ratify update cached keys or certificates immediately. They expect that images signed with compromised keys will fail validation immediately to avoid potential security attacks, and images signed with the rotated version can be validated successfully to prevent potential service downtime. To achieve this, they want to trigger Ratify to sync latest versions of the certificate and key promptly, ensuring that the necessary updates are applied. This way, images can be verified with the correct keys/certificates.

## Proposed solutions

There are two methods to update keys or certificate automatically:
- **Periodic retrieval of enabled keys or certificates**: Ratify periodically retrieve multiple enabled keys or certificates from the KMS.
  - Pros:
    - Simplicity (no need for real-time event handling).
    - Predictable resource usage.
    - Works well for less time-sensitive use cases
  - Cons:
    - May lead to delays of updates.
    - Frequent queries can impact performance.
- **Event-Driven Notification**: Ratify subscribes to KMS events, when a relevant event occurs, Ratify receives an event notification. Then Ratify can retrieve the latest enabled keys or certificates from the KMS.
    - Pros:
      - Real-time responsiveness.
      - Efficient use of resources (only fetches when needed).
      - Minimizes downtime for key or certificates changes.
    - Cons:
      - Requires adapting to different event infrastructure (webhooks, message queues).
      - Complexity in handling event delivery and retries.

The proposed solution is **Periodic retrieval of enabled keys or certificates** because it is simpler and the update is not a time-sensitive action. Users can manual update keys or certificates as a complementary if required.

As users may require previous versions for verification, so Ratify can allow users to specify how many previous versions to be synced up. Ratify's default setting is to sync two versions starting from the latest version and the previous enabled version.

## User experiences

This section describes the experience that users interact with Ratify using the proposed solution. In summary, the proposed solution maintains the existing user experience for configuring `KMP` resources. Automatic updates of keys or certificates occur seamlessly in the background. However, if users utilize inline `KMP` resources, they will still need to manually update the keys or certificates. Importantly, the automatic updates do not prevent users from making manual updates when necessary.

### Automatically update keys and certificates

If users specify key or certificate versions in the `KMP` resource, then only specific versions are updated automatically. For example, if specific versions are disabled or deleted in KMS, then they cannot be used for signature verification after updates. See an example of `KMP` configuration with the version `${KEY_VER}` specified for a key `${KEY_NAME}`:

```yaml
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

If users configure key or certificate names or aliases or IDs, no version are specified, then Ratify will sync the latest version and multiple previous versions. Ratify's default setting is to sync the latest version and one previous version. You can specify a parameter named `previousVersionCount` for multiple previous versions. The default value is `1`. If you specify the parameter value `0`, which means only latest version is synced. Disabled versions are not synced. See an example of `KMP` configuration, no versions specified for the key `${KEY_NAME}`. Ratify will fetches two versions `${KEY_VER_LATEST}` and `${KEY_VER_1}`.

```yaml
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

If your want to sync two previous versions, you can specify `previousVersionCount` to value `2`, for example,

```yaml
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
    previousVersionCount: 2
```

The following is an example of `KMP` configuration, no versions specified for the certificate `${CERT_NAME}`. Ratify will sync two certificate versions `${CERT_VER_LATEST}` and `${CERT_VER_1}`。

```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementProvider
metadata:
  name: keymanagementprovider-akv
spec:
  type: azurekeyvault
  parameters:
    vaultURI: https://${AKV_NAME}.vault.azure.net/
    certificates:
    - name: ${CERT_NAME}
    tenantID: ${TENANT_ID}
    clientID: ${IDENTITY_CLIENT_ID}
```

### Configure update interval

Users should have the ability to customize the time interval at which the retrieval process occurs, allowing them to override the default interval (e.g., 24 hours). A new parameter named `updateInterval` is introduced for a `KMP` resource using a KMS as provider, such as AKV.

The default value for `updateInterval` determines how long the updated versions of keys or certificates will be available for verifying images. For Notary Project signatures, it is the root CA certificate retrieved and configured in trust store. In most cases, root CA certificates have a long validity period. Normally certificate rotation will not result in a change on root CA certificate. For Cosign signatures with key pairs in AKV, it depends on how quickly the pipeline switching to use the latest version for signing. The update interval should also consider the impact on normal verification traffic and rotation frequency in AKV. The current recommendation for default value is 24 hours, and users can configure a proper value based on their own situations.

Below is an example to override the default retrieval interval:

```yaml
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
    updateInterval: 4h
```

### Error handling during automatic update

Some possible causes of the automated update failure are: Ratify cannot access the KMS due to permission changes, or there was a network issue during the update. If this happens, the cached keys and certificates will not be updated, and the `KMP` resource will remain the same as before. It is recommended generating warning logs for the failures and reasons, and producing failure metrics for monitoring, for example, `kmpAutoUpdateFailureCount`. The failure will not affect the next automated update. The cause of the failure could be a configuration problem on Ratify, a KMS issue, or something else. Once it is fixed, users have the option to manually update their keys and certificates instead of waiting for the next scheduled update.

### Manually update certificates and keys

You can always update existing `KMP` resources with new configuration without waiting for the automated update. To do it, you just update the existing resource file and then use `kubectl apply` command to apply the changes. For example, after you update the key name in the `KMP` resource file `my_akv_kmp.yaml`, you can execute the following command:

```shell
kubectl apply -f my_akv_kmp.yaml
```

If you do not have any updates of keys/certificates parameters in existing `KMP` resource, but you want to trigger an immediate update of keys/certificates without waiting for the next round of automated update, you can update a specific annotation in the `KMP` resource to trigger an immediate update. For example, to trigger Ratify fetches the latest version right after key rotation, the annotation `metadata.annotations.forceUpdate` is added and set to `1`. The annotation name and value are for inspiring. We can choose a better name and value handling during design.

```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementProvider
metadata:
  name: keymanagementprovider-akv
  annotations:
    forceUpdate: 1
spec:
  type: azurekeyvault
  parameters:
    vaultURI: https://${AKV_NAME}.vault.azure.net/
    keys:
    - name: ${KEY_NAME}
    tenantID: ${TENANT_ID}
    clientID: ${IDENTITY_CLIENT_ID}
```

> You can delete and recreate the resource to trigger an update. But this may cause service down time as the resource will be deleted first, then no keys or certificates can be used for verification before a new resource is created.
