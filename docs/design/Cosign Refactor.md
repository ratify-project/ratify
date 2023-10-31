# Cosign Support in Ratify 
Author: Akash Singhal (@akashsinghal)

Currently, only public registries are supported for cosign verification in Ratify. This is due to how Cosign was initially implemented and the inherent limitations with the current Ratify design on working with non-OCI Artifact manifest types. Furthermore, enabling cosign requires a flag to be set in the ORAS store. This is confusing and counter intuitive since Cosign is simply a verifier and a user should not need to interact with the Store to support a specific verifier. Even if a cosign verifier is not specified in the config, the validation process will fail for any private registry scenario if the `cosignEnabled` flag is set to true.

The goal is to redesign and implement a new Cosign integration which will remove the opt in flag from the ORAS store as well as support private registry scenarios for Cosign.

## Current Implementation

There are two main components critical to cosign verification:
- A special subroutine is added in the `ListReferrers` method that checks if the `cosignEnabled` field is set in the ORAS Store config. If cosign is enabled, an existence check for the cosign-specific tag schema manifest is performed using CRANE as the client. Currently this operation does not support any credentials to be passed in and thus fails in any auth enabled registry.
- The cosign verifier leverages cosign's online verification function to handle all pulling of image and the signature blobs and the subsequent verification operations. This violates the Referrer Store and Verifier interaction of Ratify. As a result, the cosign verifier does not have access to credentials from the ORAS auth provider. Thus, Ratify cosign verifier currently only supports public registries. 

## Goals
- Enable private registries with cosign verification
- Remove the `cosignEnabled` flag from ORAS store

The first stage will be to add support for auth enabled registries. This will involve rewriting the cosign verifier. We will need to use granular cosign verification functions with referrer store operations to verify and download the signatures instead of relying on Cosign to perform online verification. On the ORAS store side, we will switch from CRANE to ORAS for the manifest existence check that way we get auth support for "free".

The second stage will be removed the cosignEnabled flag. We'll need to stage this after since the existence manifest checks adds an extra network call that can degrade performance for notation-exclusive Ratify scenarios. Cosing recently added [support for OCI 1.1 spec](https://github.com/sigstore/cosign/pull/2684). With this cosign signatures will be discoverable via the existing `ListReferrers` funciton removing the need for the initial tag-based existence check. **However, this is an unreleased experimental feature for Cosign and thus we would need to wait before switching exclusively to this?** Once we decide to switch exclusively, we can remove the flag. For now, we can set `cosignEnabled` to true by default since now it will no longer cause errors with private registries (however it will add an extra network call and users will need to disable if performance is a concern in their scenario)

## Considerations

Previously, discussions focused on separating Cosign into a separate store. With new OCI 1.1 support, I don't believe this is worth the effort in the long run. Furthermore, even if a Cosign store was created, the credentials betwen ORAS and Cosign store would need to be shared to avoid extra overhead of rededundant authentication operations. The current auth provider implementation is used specifically with the ORAS store and the credential caching is implementd in the ORAS store too. We would need to do a fair bit of refactoring to extract the credential caching and implement a mechanism for a single auth provider to be shared between stores.