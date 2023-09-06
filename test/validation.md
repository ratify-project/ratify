# Validation
Our goal is to automate as much testing as possible with unit and integration tests. For all CLI test scenarios covered today, please see [cli-test.bats](bats/cli-test.bats). For all base ratify tests + notation tests, please see [base-test.bats](bats/base-test.bats). For all external plugin related tests, please see [plugin-test.bats](bats/plugin-test.bats). For all Azure supported test scenarios covered today, please see [azure-test.bats](bats/azure-test.bats).

## Unsupported Tests

While we are working on improving our coverage, here is the list of scenarios that currently require manual validation: 
- Azure Managed Identity Auth Provider
- AWS ECR IRSA Auth Provider

## Supported Tests

### CLI
- Verifier Scenarios
    - Notation
    - Cosign
        - Keyed
        - Keyless 
    - SBOM
    - License Checker
    - JSON Schema Validation
    - All verifier types in one
- Dynamic OCI Plugins
    - Verifier Plugin
    - Store Plugin
### Kubernetes
- Verifier Scenarios
    - Notation
    - Cosign
    - SBOM
    - License Checker
    - JSON Schema Validation
    - All verifier types in one
- ORAS Store Authentication Providers
    - Docker
    - Kubernetes Secrets
    - Azure Workload Identity
    - Azure Managed Identity
- Certificate Store Providers
    - Inline Certificate
    - Azure Key Vault Certificate
- Mutation Provider
- Dynamic OCI Plugins
    - Verifier Plugin
- CertifacteProvider CRD Status
- TLS Certificate
    - TLS Certificate Watcher
    - TLS Certificate Rotation
- High Availability Tests
    - 2 Replicas, Redis + Dapr, Notation