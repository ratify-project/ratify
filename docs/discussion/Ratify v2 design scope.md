Proposal: A more lightweight and extensible framework for Ratify v2
=======================
Author: Binbin Li

## Table of Contents
- [Proposal: A more lightweight and extensible framework for Ratify v2](#proposal-a-more-lightweight-and-extensible-framework-for-ratify-v2)
  - [Table of Contents](#table-of-contents)
  - [Background](#background)
  - [Goal](#goal)
  - [Scope](#scope)
    - [Refactoring](#refactoring)
      - [Cache Refactoring](#cache-refactoring)
      - [Policy Provider Refactoring](#policy-provider-refactoring)
      - [Configuration Refactoring](#configuration-refactoring)
      - [Reorganize the Codebase](#reorganize-the-codebase)
    - [Deprecations](#deprecations)
      - [CertificateStore](#certificatestore)
      - [Legacy Cosign Verifier](#legacy-cosign-verifier)
      - [License Checker Verifier](#license-checker-verifier)
      - [`Name` and `Type` fields in verifierReport/configuration](#name-and-type-fields-in-verifierreportconfiguration)
    - [New Features](#new-features)

## Background
Ratify has reached its first major release in over a year and has been widely adopted by users at scale. However, there are known limitations in the current design and implementation of Ratify v1 make it challenging to extend functionality. This document outlines these known limitations and proposes solutions for a more lightweight and extensible framework.

## Goal

The goal of the v2 design is to improve the current design and implementation to better meet the requirements of the customers.

1. Deprecate outdated features so that customers can have a clear understanding of the features that are still supported.
2. Refactor the code to make it more maintainable and extensible for new features.
3. Improve the performance of the system.
4. Improve the usability of the system.
5. Reduce the dependencies of the system.

## Scope

Overall, the scope of the v2 design includes the following aspects: system refactoring, deprecations, and new features. We'll discuss each aspect in detail.

### Refactoring

#### Cache Refactoring
**Current Limitation**:
1. External and internal verifiers are not sharing the same in-memory cache. This could lead to data inconsistency and performance issues. Specifically, external verifiers would create a completely new cache for each validation upon verifier initialization.
2. The OCI store behind the ORAS store is not process-safe. This could lead to race conditions when multiple external/interval verifiers are running in parallel. We already have a tracking [issue](https://github.com/ratify-project/ratify/issues/1110).
3. For the HA mode, we recommend customers to deploy Redis/Dapr as the cache backend. However, we have got feedback that the Redis/Dapr deployment along with Ratify/Gatekeeper is complex for some customers and there is concern about the availability of the Redis/Dapr.

**Proposed Solution**:
1. Remove Dapr dependency for the distributed cache backend. Ratify will access Redis directly in HA scenario.
2. Avoid race condition in the OCI store and reduce overhead for external verifiers to create new cache. There are a few options:   
    - Eliminate cache usage in external verifiers. This is the simplest solution but may impact the performance.
    - Do NOT share cache among internal and external verifiers. This is another simple solution but could not bring performance improvement.
    - Use some IPC mechanism to share cache among internal and external verifiers. For example, we can use mmap to share cache or use socket/pipe to share cache. This is more complex but could bring better performance improvement.

#### Policy Provider Refactoring
**Current Limitation**:
1. The Policy evaluation workflow is coupled within the executor. Updating or adding a new policy provider requires changes to the executor which makes it hard to maintain and extend.
2. Rego policy provider and Config policy provider return different types of results. This could lead to confusion when users are switching between different policy providers.
3. There are a few limitations/bugs in config policy that were adrresed by the Rego policy provider. And if we want to keep supporting the config policy provider, we need to make sure it is consistent with the Rego policy provider. [issue-351](https://github.com/ratify-project/ratify/issues/351), [issue-737](https://github.com/ratify-project/ratify/issues/737), [issue-528](https://github.com/ratify-project/ratify/issues/528)

**Proposed Solution**:
1. Decouple the policy evaluation workflow from the executor. This will make it easier to maintain and extend the policy providers.
2. Standardize the result format of the policy providers. This will make it easier for users to switch between different policy providers.
3. [optional] Remove the config policy provider or fix the limitations/bugs in the config policy provider.

#### Configuration Refactoring
**Current Limitation**:
1. Ratify CLI does not support all features that are available in the Ratify service. One of the reasons is that the CLI configuration does not support new options, e.g. KeyManagementProvider.
2. Oras store only supports one auth provider. One of the reasons is that the Oras store configuration does not support multiple auth providers.

**Proposed Solution**:
1. Update the CLI configuration format to be forward-compatible with missing features.
2. Update the Oras store configuration format so that it can support multiple auth providers in the future.

#### Reorganize the Codebase
**Current Limitation**:
1. We have everything in the same repo which adds the dependencies of the project. e.g. we have external plugins' implementation in the Ratify repo. 
2. The overall codebase is not well organized. We exposed some internal packages to customers and it's difficult to understand the functionality/difference of each package.

**Proposed Solution**:
1. Move the external plugins to a separate repo. This will reduce the dependencies of the project.
2. Reorganize the codebase. We should have a clear separation of the internal and external packages. We should also have a clear separation of the core/extended functionality, and CLI/Stand-alone service/K8s service.

### Deprecations
#### CertificateStore
**Current Limitation**:
1. `CertificateStore` is a previous version of `KeyManagementProvider`. It's incompatible with multi-tenancy scenario and does not support the latest features of the `KeyManagementProvider`, such as key rotation.

**Proposed Solution**:
1. Deprecate `CertificateStore` and recommend customers to use `KeyManagementProvider` instead.

#### Legacy Cosign Verifier
**Current Limitation**:
1. The legacy Cosign verifier does not support referencing multiple keys and KeyManagementProvider.
2. The legacy Cosign verifier does not support trust policy for cosign verification.

**Proposed Solution**:
1. Deprecate the legacy Cosign verifier and recommend customers to use the new Cosign verifier.

#### License Checker Verifier
**Current Limitation**:
1. The primitive implementation of the licensechecker verifier to support basic verification of license compliance (allowed license list).

**Proposed Solution**:
1. Deprecate the licensechecker verifier and recommend customers to use the new SBOM verifier.

#### `Name` and `Type` fields in verifierReport/configuration
**Current Limitation**:
1. `Name` and `Type` fields in VerifierReport refer to the name and type of the verifier that generated the report. But we got feedback that they're ambiguous and can be misleading.
2. `Name` and `Type` fields in plugin configuration are misused due to historical reasons. It makes the configuration hard to understand.

**Proposed Solution**:
1. Deprecate `Name` and `Type` fields in VerifierReport and recommend customers to use `VerifierName` and `VerifierType` instead.
2. Correct the usage of `Name` and `Type` fields in plugin configuration.

### New Features

The v2 design mainly focuses on the fundamental refactoring and deprecations that are necessary for the system. In our first prototype or first RC/official release, we may not have new features. However, we should keep the following features in mind to avoid introducing new limitations or breaking changes in the future.

1. Parity features in CLI.
    - Support KeyManagementProvider in CLI.
    - Support auth with cloud providers.
    - New command to list all configured plugins.
2. Multi-arch image validation.
3. Multi-tenancy support.
4. More policy providers and verifiers.