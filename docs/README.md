# Ratify Documentation

Within this directory you will find documentation for Ratify.  The documentation is broken down into the following sections:

## Quickstarts 

## End-to-end scenarios

- [ratify-on-aws](quickstarts/ratify-on-aws.md) - How to use ratify on AWS
- [ratify-verify-azure-cmd](quickstarts/ratify-on-azure.md) - Use notation, oras, and ratify to build and sign a container image and SBOM and verify it with the ratify cli.
- [working-with-spdx](quickstarts/working-with-spdx.md) - Use ratify with syft, oras, and verification using ratify cli.

### Plugin development:

- [creating-plugins](quickstarts/creating-plugins.md) - Details on creating your own plugins for use with ratify.

### Policy Management:

- [gatekeeper-policy-authoring](quickstarts/gatekeeper-policy-authoring.md) - Authoring gatekeeper policies for use with ratify, including rego references/examples.

### SPDX Integration:

- [working-with-spdx](quickstarts/working-with-spdx.md) - Use ratify with syft, oras, and verification using ratify cli.

### Contributing:
- [Developer Getting Started](../CONTRIBUTING.md) - Getting started with ratify development.

## Reference 

Documentation for understanding ratify and its components:

Framework and configuration:
- [Framework](reference/ratify-framework-overview.md) - The ratify framework is the core of ratify.  It is responsible for orchestrating the execution of the various plugins and providing a common interface for them to interact with each other.
- [Configuration](reference/ratify-configuration.md) - Ratify's configuration consist of:
  - [store](reference/store.md) - Ratify's store is responsible for storing and retrieving artifacts.
  - [verifier](reference/verifier.md) - Ratify's verifier is responsible for verifying the integrity of artifacts.
  - [provider](reference/providers.md) - Ratify's policy provider is used by the framework to make a final decision on if an artifact is valid or not. Policies are defined via [configuration](reference/providers.md#policy-providers).
  - [executor](reference/executor.md) - Ratify's executor is responsible for executing the plugins.


Ratify CLI
- [usage](reference/usage.md) - Additional information for using the `ratify` executable

Authentication:
- [oras-auth-provider](reference/oras-auth-provider.md) - Explanation of various authentication mechanisms available for use with ratify.

Instrumentation:
- [instrumentation](reference/instrumentation.md) - Details on the current supported instruments and metric provider setup guides

CRDs:
- [CRD references](reference/crds/) - Describes the required and optional properties of ratify CRDs

## Archive

A historical record of Ratify documentation:
