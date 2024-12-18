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
      - [Plugin Framework Refactoring](#plugin-framework-refactoring)
      - [Policy Provider Refactoring](#policy-provider-refactoring)
      - [Configuration Refactoring](#configuration-refactoring)
      - [Reorganize the Codebase](#reorganize-the-codebase)
    - [Deprecations](#deprecations)
      - [CertificateStore](#certificatestore)
      - [Legacy Cosign Verifier](#legacy-cosign-verifier)
      - [License Checker Verifier](#license-checker-verifier)
      - [`Name` and `Type` fields in verifierReport/configuration](#name-and-type-fields-in-verifierreportconfiguration)
    - [Documentation Update](#documentation-update)
    - [Unit Test Improvement](#unit-test-improvement)

## Background
Ratify has reached its first major release in over a year and has been widely adopted by users at scale. However, there are known limitations in the current design and implementation of Ratify v1 make it challenging to extend functionality. This document outlines these known limitations and proposes solutions for a more lightweight and extensible framework.

## Goal

The goal of the v2 design is to improve the current design and implementation to better meet the requirements of the customers.

1. Streamline our offerings by phasing out outdated features, ensuring customers have a clear view of all actively supported features.
2. Refactor the code to make it more maintainable and extensible for new features.
3. Improve the performance of the system.
4. Improve the usability of the system.
5. Reduce the dependencies of the system.

## Scope

Overall, the scope of the v2 design includes the following aspects: system refactoring, deprecations, document update and unit test improvement. We'll discuss each aspect in detail.

Notes: The priority of each task is scored from 1 to 5, with 1 being the highest priority.

### Refactoring

#### Plugin Framework Refactoring
**Current Limitation**:
1. External and internal verifiers are not sharing the same in-memory cache. Each external verifier will have a separate ReferrerStore while created. This could lead to data inconsistency and performance issues.
2. The OCI store behind the ORAS store is not process-safe. This could lead to race conditions when multiple external/interval verifiers are running in parallel. We already have a tracking [issue](https://github.com/ratify-project/ratify/issues/1110).
3. The built-in plugins are tightly coupled with the Ratify core. Plugin version upgrade requires a new Ratify release. And built-in plugins also add dependencies to the Ratify core which makes it hard to maintain.

**Proposed Solution**:
1. We need to refactor the whole plugin framework to make it more lightweight and concurrent-safe. Specifically, we have 2 options:
   1. Make all plugins as external plugins. Plugins can communicate between each other through IPC mechanism such as RPC and pipe. This can make the plugins decoupled from the Ratify core logic and make it easier to maintain and extend. However, this may introduce some overhead of IPC communication and data serialization/deserialization. e.g. [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
   2. Make plugins run in the same process as Ratify main process. There are 2 options, one is to use `plugin` package in Go to build .so files and load them dynamically. The other is to make plugins as libraries and import them dynamically during build time. This can make the plugins behave like built-in plugins but with less dependencies in the Ratify main repo.
2. We may need to create separate repos for plugins to decouple the dependencies from the Ratify core.
**Priority**: The plugin framework refactoring is priority 1 as it's the fundamental part of the project.

#### Policy Provider Refactoring
**Current Limitation**:
1. The Policy evaluation workflow is coupled within the executor. Updating or adding a new policy provider requires changes to the executor which makes it hard to maintain and extend.
2. Rego policy provider and Config policy provider return different types of results. This could lead to confusion when users are switching between different policy providers.
3. There are a few limitations/bugs in config policy that were addressed by the Rego policy provider. And if we want to keep supporting the config policy provider, we need to make sure it is consistent with the Rego policy provider. [issue-351](https://github.com/ratify-project/ratify/issues/351), [issue-737](https://github.com/ratify-project/ratify/issues/737), [issue-528](https://github.com/ratify-project/ratify/issues/528)

**Proposed Solution**:
1. Decouple the policy evaluation workflow from the executor. This will make it easier to maintain and extend the policy providers.
2. Standardize the result format of the policy providers. This will make it easier for users to switch between different policy providers.
3. Make the rego policy provider as the default policy provider and provide some templates covering most user scenarios.
4. [optional] Remove the config policy provider or fix the limitations/bugs in the config policy provider.

**Priority**: The overall policy provider refactoring is priority 2. But if we want to fix the limitations/bugs in the config policy provider, it can be a task of priority 3.

#### Configuration Refactoring
**Current Limitation**:
1. Ratify CLI does not support all features that are available in the Ratify service. One of the reasons is that the CLI configuration does not support new options, e.g. KeyManagementProvider.
2. The K8s ratify service configuration consists of multiple different CRDs, where Verifiers will also reference KMPs. This makes it hard for beginners to configure and troubleshoot the system. And some users also reported that they just need to ConfigMap to configure the service without touching CRDs.
3. It's not straightforward to convert configurations between CLI and service. Users may need extra effort to learn the difference between CLI and service configurations.
4. Oras store only supports one auth provider. One of the reasons is that the Oras store configuration does not support multiple auth providers. Related issue: [issue-974](https://github.com/ratify-project/ratify/issues/974).

**Proposed Solution**:
1. Refactor both CLI and service configurations to make them consistent and convertible. We can have a common configuration format that can be used by both CLI and service. We hope to make the new configuration format extensible to standalone service and GRPC service.
2. Simplify the configuration options. Remove the unnecessary options and make the configuration more user-friendly.
3. Update the Oras store configuration format so that it can support multiple auth providers in the future.

**Priority**: Ratify CLI config refactoring can be priority 1 as it's the fundamental part for CLI. Oras store config refactoring can be priority 2 or 3 as it's not a common scenario to have multiple auth providers.

#### Reorganize the Codebase
**Current Limitation**:
1. Currently Ratify only supports CLI and Gatekeeper add-on user scenarios. In the future, we may also supports other scenarios including standalone service, containerD plugin, and github action. But the current codebase have everything in the same repo which makes it hard to maintain and extend.
2. We have everything in the same repo which adds the dependencies of the project. e.g. we have external plugins' implementation in the Ratify repo. 
3. The overall codebase is not well organized. We exposed some internal packages to customers and it's difficult to understand the functionality/difference of each package.

**Proposed Solution**:
1. Move plugins to separate repos. This will reduce the dependencies of the project.
2. Reorganize the codebase. Follow the best practices of Go project layout.
3. Extract out the Ratify core and different user scenarios to different repos, e.g. ratify-cli, ratify, ratify-containerd.

**Priority**: Creating different repos is the first step while refactoring the codebase, therefore it's priority 1. Reorganize the codebase can be done later with priority 2.

### Deprecations
#### CertificateStore
**Current Limitation**:
1. `CertificateStore` is a previous version of `KeyManagementProvider`. It's incompatible with multi-tenancy scenario and does not support the latest features of the `KeyManagementProvider`, such as key rotation.

**Proposed Solution**:
1. Deprecate `CertificateStore` and recommend customers to use `KeyManagementProvider` instead.

**Priority**: 1

#### Legacy Cosign Verifier
**Current Limitation**:
1. The legacy Cosign verifier does not support referencing multiple keys and KeyManagementProvider.
2. The legacy Cosign verifier does not support trust policy for cosign verification.

**Proposed Solution**:
1. Deprecate the legacy Cosign verifier and recommend customers to use the new Cosign verifier.

**Priority**: 1

#### License Checker Verifier
**Current Limitation**:
1. The primitive implementation of the licensechecker verifier to support basic verification of license compliance (allowed license list).

**Proposed Solution**:
1. Deprecate the licensechecker verifier and recommend customers to use the new SBOM verifier.

**Priority**: 1

#### `Name` and `Type` fields in verifierReport/configuration
**Current Limitation**:
1. `Name` and `Type` fields in VerifierReport refer to the name and type of the verifier that generated the report. But we got feedback that they're ambiguous and can be misleading.
2. `Name` and `Type` fields in plugin configuration are misused due to historical reasons. It makes the configuration hard to understand.

**Proposed Solution**:
1. Deprecate `Name` and `Type` fields in VerifierReport and recommend customers to use `VerifierName` and `VerifierType` instead.
2. Correct the usage of `Name` and `Type` fields in plugin configuration.

**Priority**: 1

### Documentation Update
Post or along with the v2 release, we should update the documentation to reflect the changes in the v2 design. The documentation update should include the following aspects:
1. Announce deprecations and provide migration guides.
2. Update the architecture and design documents.
3. Update the user guide and CLI reference.
4. Update the developer guide and plugin development guide.

**Priority**: As we have received a few feedback that the current documentation is not up-to-date, the documentation update should be priority 1.

### Unit Test Improvement
We should improve the unit test coverage and quality to ensure the stability of the system. The current test coverage of Ratify repo is around 72%, and we noticed that it's quite difficult for contributors to add/update unit tests that cover some legacy code. We should improve the test coverage and quality by:
1. Refactor the legacy code to make it more testable.
2. Add more unit tests to cover the new features.

Actually the refactoring tasks mentioned above can be considered as the unit test improvement tasks as well.

**Priority**: 2