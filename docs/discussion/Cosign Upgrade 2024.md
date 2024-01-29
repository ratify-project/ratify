# Cosign Upgrade 2024
Author: Akash Singhal (@akashsinghal)

Tracked issues in scope:
- [Support Cosign verification with multiple keys](https://github.com/deislabs/ratify/issues/1191)
- [Support for Cosign verification with keys managed in KMS](https://github.com/deislabs/ratify/issues/1190)
- [Support Cosign verification with RSA key](https://github.com/deislabs/ratify/issues/1189)

Ratify currently supports keyless cosign verification which includes an optional custom Rekor server specification. Transparency log verification only occurs for keyless scenarios. Keyed verification is limited to a single public key specified as a value provided in the helm chart. The chart creates a `Secret` for the cosign key and mounts it at a well-known path in the Ratify container. Users must manually update the `Secret` to update the key. There is no support for multiple keys. There is no support for keys stored KMS. There is only support for ECDSA keys, and not RSA or ED25519. There is no support for certificates.

## Support key configuration as K8s resource

Currently, cosign verifier looks for a single key that already exists at a specified path in the Ratify container. Ratify helm chart contains a `Secret` which is mounted at the specified mount point. Ratify's key management experience should be decoupled from secrets and mount paths. Instead, it should be a first class key management experience similar to how certificates are managed via `CertificateStore`.

### Option 1: Extend existing Certificate Store

#### User Experience

Inline certificate provider with key
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: CertificateStore
metadata:
  name: ratify-cosign-inline-key
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: inline
  parameters:
    type: key
    value: |
    ---------- BEGIN RSA KEY ------------
    ******
    ---------- END RSA KEY ------------
    
```

Azure Key Vault certificate provider with keys
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: CertificateStore
metadata:
  name: certstore-akv
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: azurekeyvault
  parameters:
    vaultURI: VAULT_URI
    keys:
      - keyName: KEY_NAME
        keyVersion: KEY_VERSION
    tenantID: TENANT_ID  
    clientID: CLIENT_ID
```

Azure Key Vault certificate provider with keys + certificates
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: CertificateStore
metadata:
  name: certstore-akv
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: azurekeyvault
  parameters:
    vaultURI: VAULT_URI
    keys:
      - keyName: KEY_NAME
        keyVersion: KEY_VERSION
    certificates:
      - certificateName: CERTIFICATE_NAME
        certificateVersion: CERTIFICATE_VERSION
    tenantID: TENANT_ID  
    clientID: CLIENT_ID
```
#### Implementation Details

- New API `GetKeys`
  - added to the `CertificateStore` API
  - return an array of `Key`s. 
    - contains `byte` array content
- Global `Key`s map
  - follow same pattern as `Certificates` map which is updated on reconcile of the `CertificateStore` K8s resource.
- Can we promote Cosign to be a built-in verifier like Notation is?
  - This would allow us to use `CertificateStore` without having to build support for external plugins accessing certificates.
  - It will also be slightly more performant if it can share the in-memory ORAS store cache.
  - **UPDATE 1/18/24: Cosign verifier will be built-in**
- ~~How do we share certificates and keys with external plugins?~~
  - ~~All certs are loaded in an in-memory map currently. The map cannot be shared with external processes.~~
  - ~~External plugins should NOT be pulling certificates/keys per invocation (external KMS would be stressed. result in very expensive calls)~~
  - ~~Option 1: Serialize and add `Certificates` & `Keys` map into the external plugin json input~~
    - ~~Embed each map during plugin invocation~~
      - ~~namespace context aware for multi-tenancy~~
      - ~~Is security a concern for embedding certs + keys to invoke plugins not related to signature verification?~~
    - ~~Cosign verifier will be updated to deserialize and consume the maps~~
  - ~~Option 2: Store certs + keys on disk at well-known path~~
    -  ~~`CertificateStore` will be refactored to store fetched certs from providers in a directory on disk~~
      ```diff
      - .ratify/certificate_store/
      - <provider_name>-<namespace>/
      -   keys/
      -     key1.rsa
      -   certificates/
      -     cert1.pem
      -     cert2.pem
      ```
    - ~~Configuration loading from JSON (CLI) or CRD reconcile loop will be responsible for deleting the directories of cert providers as they are removed~~
    - ~~External plugins will have to have the `CertificateStore` config passed in as JSON input in order to create instance of the cert provider objects~~
    - ~~External plugins will need to be namespace aware in order to know which directory to access~~
    - ~~Pros:~~
      - ~~Cosign plugin can manage accessing its own keys + certificates without executor having to coordinate~~
      - ~~Other external plugins will not get unnecessary keys + certs provided in json input~~ 
    - ~~Cons:~~
      - ~~Keys + certs must be added/deleted off disk~~
        - ~~What happens if keys + certs are left on disk due to Ratify failure?~~
        - ~~Keys + certs will be able to be accessed by any process with filesystem access~~
        - ~~CLI scenario will require fetching + deleting certs + keys off disk per invocation~~
      - ~~Existing cert providers used by notation will need to be updated to read/write certs from disk~~
      - ~~External plugins will need to be aware of namespace context in order to only access resources in namespaced directory~~
  - ~~Option 3 (NOT recommended): Add certificates + keys to shared cache~~
    - ~~Define two new cache keys partitioned by namespace~~
    - ~~CANNOT use the ristretto in-memory cache as it is since external cosign process must be able to access it~~
      - ~~This would require hard dependency on HA mode which is a bad experience. Why should a user need HA distributed cache just for Cosign?~~
      - ~~Implement a new cache provider for cache sharing between processes on same host~~
        - ~~NOT trivial~~
        - ~~Require a new http cache server and clients~~
    - ~~Cache provider initialization in cosign plugin implementation~~
    - ~~One advantage is that not only will cert + key sharing be solved but also oras store credentials can easily shared which is currently not possible for external plugins~~
- Extend Inline Certificate Provider
  - add new `type` field to spec: "key" or "certificate". If not provided, default to "certificate" for backwards compatability. 
- Extend Azure Key Vault Certificate Provider
  - use `GetKeys` API to fetch keys from configured key vault
  - add new `keys` field to spec
- Multitenancy considerations
  - Need a separate keys map for namespaced and global certificate store keys

### **Option 2: Renaming CertificateStore to KeyManagementSystem** SELECTED IMPLEMENTATION
- Same as Option 1 except we introduce a new CRD that will replace `CertificateStore` due to potential naming confusion
- Is the existing name confusing enough to warrant changing to KeyManagementSystem?
- How would we support CRD name change?
  - Introduce new `KeyManagementSystem` CRD resource and then deprecate `CertificateStore`

#### Proposed Config Changes

Compared to the `CertificateStore`, the `KeyManagementSystem` config spec could be updated to be more flexible. A new `name`` field will be used only in CLI scenarios to mirror CRD name functionality as a unique identifier. This enables multiple KMS of same type to be used.

```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: KeyManagementSystem
metadata:
  name: ratify-notation-inline-cert-kms
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  name: inline
  parameters:
    contentType: key
    value: |
      ---------- BEGIN RSA KEY ------------
      ******
      ---------- END RSA KEY ------------
```

CLI Config
```json
{
  ...
  "keyManagementSystems": {
    {
      "name": "ratify-notation-inline-cert-kms",
      "provider": "inline",
      "contentType": "key",
      "value": "---------- BEGIN RSA KEY ------------
      ******
      ---------- END RSA KEY ------------"
    },
    {
      "name": "ratify-notation-inline-cert-kms-2",
      "provider": "inline",
      "contentType": "key",
      "value": "---------- BEGIN RSA KEY ------------
      ******
      ---------- END RSA KEY ------------"
    },
  }
  ...
}
```

**Do we want to consider making the inline provider accept an array of content?**

### Option 3: Introduce new CRD `KeyStore` separate from `CertificateStore`

All of the same considerations as Option 1. It will be a completely new CRD with an inline and AKV provider equivalent.

Pros: 
- More intuitive naming: embedding 'key' specification in cert store might be confusing
- No breaking changes or backwards compatability requirements
Cons:
- Yet one more resource type user must be concerned to configure
- Providers will share almost the exact same implementation as certificate stores

### How do we support CLI scenario?
  - `CertificateStore` is only supported in K8s scenarios as CRD. There is no CLI equivalent.
  - Ideally, `CertificateStore` should be supported at CLI too. However, even with CLI mode, a local directory for keys should be an option user can specify rather than just in-line and AKV.
  - Cosign plugin will handle directory configuration and key reading within plugin implementation. This should be complimentary to any certs provided as plugin input.
  - Should we have a separate local directory cert store provider that gets configured like other cert stores or should we scope it only to the cosign plugin?
    - In K8s scenarios, it will not be an encouraged pattern to specify local directory. This is why implementing directory cert reading as a cert store provider may not make  sense.
    - However, if a user decides to do a mixture of cert stores + local directory then they would need to specify in 2 different sections (cosign plugin config + cert store)
    - **UPDATE 1/18/24**: We will mirror notation experience with a `verificationKeys` field in the cosign verifier config

## Support Multiple keys for cosign verification

Currently, there is only support for a single key per cosign verifier. 

- Goals:
  - Cosign verifier should be updated to trust a set of keys and perform verification using that trusted key set
  - Cosign verifier MUST match the verification logic of Cosign CLI for a single Key
  - Cosign verifier should be able to operate a loose ('All keys are equally trusted and ONLY one is requred to be used in validation') or strict (ALL keys are trusted and ALL keys must be used in validation) enforcement.

### How does Cosign CLI verification work today?

- Image manifest pushed to same repository as subject (tag schema: `sha256-abc123.sig`)
- Image manifest is updated every time new signature is added. Each layer is a different signature (possibly from different signing keys)
```json
# Sample Cosign OCI Image Manifest
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 451,
    "digest": "sha256:bbacdc3dc4b33c868c6c27192cb1d1bda619c1454abb57ebea806eac13250d3c"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 254,
      "digest": "sha256:75a45fb043a8abfe52f256cb1a424d8808e886a3f40886228cdfa9b8d75c805d",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEQCIGoZazNt1/eEBc08AmcqXY+yNQQJY2ziZPiIopDjHnU5AiAUgeWBy90GgqDdnWpxd1b3lbqzgr2UV2ZyR59eDOyxVg=="
      }
    },
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 254,
      "digest": "sha256:75a45fb043a8abfe52f256cb1a424d8808e886a3f40886228cdfa9b8d75c805d",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEUCIBVcGGZR6pvJ2SQ8KI0TZY02SLFSSOFJuJph2Qcm8cWZAiEArL+BF5BGRGxELVBeJtAUy8i7L1uuBeUT9fF6TrFYVLQ="
      }
    },
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 254,
      "digest": "sha256:75a45fb043a8abfe52f256cb1a424d8808e886a3f40886228cdfa9b8d75c805d",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEQCIEIsK5yr1kGS1dgRmbudT55oghADdyAOfUwlnayPKAmmAiBTdZkhgHkZdw11tullz7HNOg5O1r5U/j4m2RE39b92Vg=="
      }
    }
  ]
}
```
- Image signature succeeds if AT LEAST one of the signatures verifies true with the corresponding key
  - Validation occurs on each signature in the manifest in parallel against the SAME key
  - Cosign only supports validation against a SINGLE key at a time

### Trust Policy
Cosign verifier should support multiple trust policies based on the KeyManagementSystems (KMS) enabled and the desired verification scenario. Please refer to [this](#user-scenarios) section for common user scenarios. At a high level users must be able to have:
  - multiple KMS `inline` key resources (each `inline` will have a single key)
  - multiple keys defined in a single AKV KMS
  - multiple KMS `inline` certificate resources (each `inline` may have a single cert or a cert chain)
  - multiple certificates definine in a single AKV KMS
  - mix of keys + certificates in a single AKV KMS
  - a way to specify specific keys/certificates within a KMS by name and version
  - a way to scope a policy to a particular registry/repo
  - multiple policies depending on registry scope
  - a way to specify enforcement for that specific policy
    - skip: don't perform any verification for an image reference that matches this policy
    - any: at least one of the keys/certificates trusted in the policy must result in a successful verification for overall cosign verification to be true
    - all: ALL keys/certificates trusted in the policy must result in a successful verification for overall cosign verification to be true

Sample
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: verifier-cosign
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  name: cosign
  artifactTypes: application/vnd.dev.cosign.artifact.sig.v1+json
  parameters:
    trustPolicies:
      - name: wabbit-networks-images
        registryScopes:
          - wabbitnetworks.myregistry.io/carrots
          - wabbitnetworks.myregistry.io/beets
        keys: # list of keys that are trusted. Only the keys in KMS are considered
          - kms: inline-keys-1 # REQUIRED: if name is not provided, all keys are assumed to be trusted in KeyManagementSystem resource specified
          - kms: inline-keys-2
          - kms: akv-wabbit-networks
            name: wabbit-networks-io # OPTIONAL: key name
            version: 1234567890 # OPTIONAL: key version (inline will not support version)
        certificates: # list of certificates that are trusted. Only the certificates in KMS are considered
          - kms: inline-certs-1
        enforcement: any # skip (don't perform any verification and auto pass), any (at least one key/cert used in successfull verification is overall success), all (all keys/certs must be used for overall success)
      - name: verification-bypass
        registryScopes:
          - wabbitnetworks.myregistry.io/golden
        enforcement: skip
```

To start, only a single `trustPolicies` entry + `keys` with `kms`, `name`, and `version` will be supported. The behavior will match an equivalent `enforcement` of `any`. Future, work will add support for remainig configs.
### User Scenarios

#### 1 Signature, 2 trusted keys
Bob has a container image he built. His trusted pool contains 2 self-managed keys to sign the image using `cosign`. Only ONE of the keys is used for signing at a time. There is only ONE signature. Bob's organization utilizes and trusts both keys. Bob wants to ensure all container images entering his K8s cluster are verified to have a valid cosign signature using AT LEAST one key from a trusted pool.

- Bob installs Ratify on the cluster
- Bob applies 2 new inline `KeyManagementSystem` CR onto the cluster, `inline-key-1` & `inline-key-2`. Each CR will have a key Bob's organization trusts.
```yaml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: KeyManagementSystem
metadata:
  name: inline-key-1
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: inline
  parameters:
    type: key
    value: |
    ---------- BEGIN RSA KEY ------------
    ******
    ---------- END RSA KEY ------------
    
```
```yaml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: KeyManagementSystem
metadata:
  name: inline-key-2
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: inline
  parameters:
    type: key
    value: |
    ---------- BEGIN RSA KEY ------------
    ******
    ---------- END RSA KEY ------------
    
```
- Bob applies the cosign verifier CR
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: verifier-cosign
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  name: cosign
  artifactTypes: application/vnd.dev.cosign.artifact.sig.v1+json
  parameters:
    trustPolicies:
      - name: multiple-trusted-keys
        keys:
          - kms: inline-key-1
          - kms: inline-key-2
```
- Bob attempts to deploy a pod from an image that has cosign signature signed with key in KeyManagementSystem `inline-key-1`. Pod is verified and created successfully.
- Bob attempts to deploy a pod from an image that has cosign signature signed with key in KeyManagementSystem `inline-key-2`. Pod is verified and created successfully.
- Bob attempts to deploy a pod from an image that has cosign signature signed with key NOT in `inline-key-1` or `inline-key-2`. Pod FAILS verification and blocked.

#### 2 Signatures but only 1 is from a key that he trusts

Bob has a container image that he imported from another registry. An existing cosign signature, signed by an entity he doesn't trust, is already associated with the image. After vetting the image, he utilizes 1 self-managed key to sign the image using `cosign`. Now the image has 2 cosign signatures. Bob only trusts his key. Bob wants to ensure all container images entering his K8s cluster are verified to have a valid cosign signature using only HIS key that he trusts.

- Bob installs Ratify on the cluster
- Bob applies a new inline `KeyManagementSystem` CR onto the cluster, `inline-key`
```yaml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: KeyManagementSystem
metadata:
  name: inline-key
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: inline
  parameters:
    type: key
    value: |
    ---------- BEGIN RSA KEY ------------
    ******
    ---------- END RSA KEY ------------
    
```

- Bob applies the cosign verifier CR
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: verifier-cosign
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  name: cosign
  artifactTypes: application/vnd.dev.cosign.artifact.sig.v1+json
  parameters:
    trustPolicies:
      - name: single-trusted-key
        keys:
          - kms: inline-key
```
- Bob attempts to deploy a pod from the vetted image that has 2 cosign signatures, one of which is signed with key in KeyManagementSystem `inline-key`. Pod is verified and created successfully.
- Bob attempts to deploy a pod from an image that has cosign signature(s) signed with a different key in `inline-key`. Pod FAILS verification and blocked.

#### 2 Signatures 2 Keys: both keys must be used

Bob has a container image that is produced from a build pipeline and tested via a testing pipeline. After each pipeline, the image is signed with cosign using a SEPARATE key. Now the image has 2 cosign signatures. Bob trusts both keys and requires BOTH keys to be used. Bob wants to ensure all container images entering his K8s cluster are verified to have a valid cosign signature from his build AND test pipeline.

- Bob installs Ratify on the cluster
- Bob applies a new inline `KeyManagementSystem` CR onto the cluster, `inline-key-build` & `inline-key-test`
```yaml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: KeyManagementSystem
metadata:
  name: inline-key-build
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: inline
  parameters:
    type: key
    value: |
    ---------- BEGIN RSA KEY ------------
    ******
    ---------- END RSA KEY ------------
    
```
```yaml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: KeyManagementSystem
metadata:
  name: inline-key-test
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  provider: inline
  parameters:
    type: key
    value: |
    ---------- BEGIN RSA KEY ------------
    ******
    ---------- END RSA KEY ------------
    
```

- Bob applies the cosign verifier CR
```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: verifier-cosign
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "5"
spec:
  name: cosign
  artifactTypes: application/vnd.dev.cosign.artifact.sig.v1+json
  parameters:
    trustPolicies:
      - name: build-test-verification
        keys:
          - kms: inline-key-build
          - kms: inline-key-test
        enforcement: all
```
- Bob attempts to deploy a pod from the vetted image that has 1 cosign signature, which is signed with key in KeyManagementSystem `inline-key-build`. Pod FAILS verification and blocked.
- Bob attempts to deploy a pod from the vetted image that has 2 cosign signatures, which is signed with keys in KeyManagementSystem `inline-key-build` & `inline-key-test`. Pod passes verification and is created.

> **NOTE**: Based on Cosign's own verification implementation, there does not seem to be a way to enforce that ALL signatures found are valid. This behavior is unchartered as far as we can tell and would require Ratify to determine if we want to support this directly. For that reason, the strict ALL signature scenario is not yet included.

#### Extra Considerations
- How does Kyverno handle multiple Keys?
  - Each key specified is a separate cosign verification operation
  - User specifies enforcement type by specifying `count` of verification operations that must succeed.
    - e.g 2 keys specified and 2 signatures on the image. each verification is against both signatures for a single key
    - If count is 2, then 2 verifications must return true
- How will this work with OCI 1.1 enabled cosign?
  - Cosign OCI 1.1 support is experimental
  - Currently, an arbitrary signature from list of referrers is chosen to perform verifications on IF multiple cosign artifacts are found.
  - This logic will be updated in the future once 1.1 is GA but for now it's left as a TODO in the Cosign code base.

### Implementation Details

#### Handling multiple verifications at once
- Ratify will largely mirror how cosign performs multi-signature verification
  - cannot invoke cosign's multi signature verification API since the API exposed also includes the registry content fetch which is independent in Ratify
- separate go routines per signature and collect all results PER KEY 

## Support RSA and ED25519 keys for verification

Each key specified will encapsulate a single verification context for all signatures found in the artifact that the verifier is running for.

Verification flow per key:
- Take raw key content (byte array) and [`LoadPublicKeyRaw()`](https://github.com/sigstore/cosign/blob/daf1eeb5ff2022d11466b8eccda25768b2dd2992/pkg/signature/keys.go#L91C1-L91C1)
- The `LoadPublicKeyRaw()` method in the cosign keys library will automatically detect which key type (based on the key header) and return the appropriate signature verifier
- Verifier is then passed to cosign's `VerifySignature()` method as a verification option

## Dev Work Items (WIP)
- Introduce new `KeyManagementSystem` CRD to replace `CertificateStore` (~ 3 weeks)
  - maintain old `CertificateStore` controllers and source code for backwards compat
  - deprecate CRD version for `CertificateStore`
  - define new `KeyManagementSystem` CRD + controllers
  - port certificate providers implementation to new `KMS` object
  - refactor to factory paradigm
  - refactor to define rigid config schema (currently, only a generic map of attributes passed)
  - add plugin support
- Add Key support to `KeyManagementSystem` (~ 2 weeks)
  - update API
  - update Inline provider with `type` field
  - update AKV provider for `key` fetching logic
  - update controllers to add new key maps for global + namespaced keys
  - update controllers to add new certs maps for namespaced certs
- ~~Update key/certificate store logic to provide cert + key access to external plugins (see [above](#implementation-details))~~
  - ~~Option 1 or 2 depending on what's decided~~
  - ~~NOTE: If we make cosign a built-in verifier, we will NOT have to do this.~~
- Cosign verifier multi key support (~ 2.5 weeks)
  - add `trustPolicies`
    - support multiple `keys` each with `kms`, `name`, and `version`
  - add multi key verification logic including concurrent signature verification using routines
- Add RSA and ED25519 key support (~ 0.5 week)
  - auto detect key type based on parsing library
  - pick cosign verifier according to format
- Add docs and walkthroughs (~ 1 week)
  - redo cosign walk through
- New e2e tests for different scenarios (~ 1 week)
- Add `KeyManagementSystem` support to CLI (~ 1.5 weeks)

## Future Considerations
- Support `registryScopes` for repo based `trustPolicies`
- Support `enforcement` with `skip`, `any`, and `all`
- Support attestations with cosign signature embedded
- Support certificate based verification of cosign signatures
- Verify image annotations
- Custom signature repository
- Transparency log and SCT verification options
- Keyless signature verification options
  - Custom Rekor server
  - Github workflows