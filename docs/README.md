# Ratify Documentation

This is an overview of the documentation available today for Ratify in the folders

Titles by sub-folder in this directory

## examples

These walk through end-to-end examples of using ratify.

- [ratify-on-aws](examples/ratify-on-aws.md) - How to use ratify on AWS
- [ratify-verify-azure-cmd](examples/ratify-verify-azure-cmd.md) - Use notation, oras, and ratify to build and sign a container image and SBOM and verify it with the ratify cli.
- [working-with-spdx](examples/working-with-spdx.md) - Use ratify with syft, oras, and verification using ratify cli.

## reference

These documents can be found generally useful to understand using or implementing ratify, but do not walk through end-to-end examples.

- [gatekeeper-policy-authoring](reference/gatekeeper-policy-authoring.md) - Authoring gatekeeper policies for use with ratify when it is in passthrough execution mode, including rego references/examples.
- [oras-auth-provider](reference/oras-auth-provider.md) - Explanation of various authentication mechanisms available for use with ratify.

## developer

The documents in this directory are for developers who want to contribute to Ratify or want to understand the internals of Ratify.

- [Contributing](../CONTRIBUTING.md) - How to get ratify development environment setup and generally contribute back to Ratify.
- [README](./developer/README.md) - breaks down the architecture of Ratify and how it works.
- [providers](./developer/providers.md) - information about built-in providers and the extensible policy provider interface
- [executor](./developer/executor.md) - information about the executor and how it works
- [store](./developer/store.md) - information about the store plugin and how it works
- [verifier](./developer/verifier.md) - information about the verifier plugin and how it works.

### docs improvements backlog

developer

- Overhaul the root readme as this was from original creation and the design has been more finalized since then.  Likely combined effort with the next item to streamline.
- Streamline documentation between contributing, framework readme and providers, executor, store, verifier docs.
- [Guidance for new plugins](https://github.com/deislabs/ratify/issues/405)
- [Create a new plugin scaffold](https://github.com/deislabs/ratify/issues/8)

examples

- [Azure e2e walkthrough](https://github.com/deislabs/ratify/issues/59)
- [Cosign walkthrough](https://github.com/deislabs/ratify/issues/230)
- Using ratify in pass-through execution mode
