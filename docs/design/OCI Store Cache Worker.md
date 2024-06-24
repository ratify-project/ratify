# OCI Store Index Race Condition and Cache Worker

## Problem Statement

Tracked issues in scope:
<https://github.com/ratify-project/ratify/issues/1110>

ORAS referrer store can fail to initialize for external verifiers with error:

```log
Original Error:(Original Error: (create store from input config failed with error Original Error: (invalid OCI Image Index: failed to decode index file: EOF), Error: plugin init failure, Code: PLUGIN_INIT_FAILURE, Component Type: referrerStore, Detail: could not create local ORAS cache at path: /home/runner/.ratify/local_oras_cache))
```

This occurs when multiple external verifiers are executed for the same subject in parallel.
Intermittent issue affecting e2e tests

```json
{
  "verifierReports": [
    {
      "subject": "localhost:5000/all:v0",
      "isSuccess": true,
      "name": "notation",
      "message": "signature verification success",
      "extensions": {
        "Issuer": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US",
        "SN": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US"
      },
      "artifactType": "application/vnd.cncf.notary.signature"
    },
    {
      "isSuccess": false,
      "name": "cosign",
      "message": "Original Error: (Original Error: (create store from input config failed with error Original Error: (invalid OCI Image Index: failed to decode index file: EOF), Error: plugin init failure, Code: PLUGIN_INIT_FAILURE, Component Type: referrerStore, Detail: could not create local oras cache at path: /home/runner/.ratify/local_oras_cache), Error: verify signature failure, Code: VERIFY_SIGNATURE_FAILURE, Plugin Name: cosign, Component Type: verifier), Error: verify reference failure, Code: VERIFY_REFERENCE_FAILURE, Plugin Name: cosign, Component Type: verifier",
      "artifactType": "application/vnd.dev.cosign.artifact.sig.v1+json"
    },
    {
      "subject": "localhost:5000/all:v0",
      "isSuccess": true,
      "name": "licensechecker",
      "message": "License Check: SUCCESS. All packages have allowed licenses",
      "artifactType": "application/vnd.ratify.spdx.v0"
    },
    {
      "subject": "localhost:5000/all:v0",
      "isSuccess": true,
      "name": "schemavalidator",
      "message": "schema validation passed for configured media types",
      "artifactType": "vnd.aquasecurity.trivy.report.sarif.v1"
    },
    {
      "isSuccess": false,
      "name": "sbom",
      "message": "Original Error: (Original Error: (create store from input config failed with error Original Error: (invalid OCI Image Index: failed to decode index file: EOF), Error: plugin init failure, Code: PLUGIN_INIT_FAILURE, Component Type: referrerStore, Detail: could not create local oras cache at path: /home/runner/.ratify/local_oras_cache), Error: verify signature failure, Code: VERIFY_SIGNATURE_FAILURE, Plugin Name: sbom, Component Type: verifier), Error: verify reference failure, Code: VERIFY_REFERENCE_FAILURE, Plugin Name: sbom, Component Type: verifier",
      "nestedResults": [
        {
          "subject": "localhost:5000/all@sha256:b71c1f874fbc92173278bcb7bb44c785b167f7efa3c44b52eb48e20d540741b5",
          "isSuccess": true,
          "name": "notation",
          "message": "signature verification success",
          "extensions": {
            "Issuer": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US",
            "SN": "CN=ratify-bats-test,O=Notary,L=Seattle,ST=WA,C=US"
          },
          "artifactType": "application/vnd.cncf.notary.signature"
        }
      ],
      "artifactType": "org.example.sbom.v0"
    }
  ]
}
```

## User Scenarios

### Using multi-verifier

1. Multiple different verifiers are executed for the same subject
2. Multiple same verifiers with different version executed for the same subject

### Multi-tenancy

1. reusing OCI artifacts

## Purpose

- Resolve OCI Store race condition in multi-verifier scenarios
- Ensure deduplication of OCI artifacts hold in local environment

## Data Flow

```mermaid
%% Initialize diagram configuration
%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#ffcc00', 'edgeLabelBackground':'#f8f8f8', 'tertiaryColor': '#fff'}}}%%
flowchart TD
    %% Define styles
    classDef subProcess fill:#ffcc00,stroke:#333,stroke-width:2px;

    %% Define components
    subgraph VerifierSubsystem [Verifier Subsystem]
        GetSubjectDescriptor[Get Subject Descriptor]
        GetSubjectReferenceManifest[Get Subject Reference Manifest]
        GetReferrerArtifact[Get Blob Content]
        PerformVerification[Perform Verification]
        ReturnVerifierResult[Return Verifier Result]
        MemoryBasedStore[Memory Based Store]
        OCIStoreCache[OCI Store Cache]
        OCIStoreCacheWorker[Cache Worker]
        CacheProvider[Cache Provider]
        ORASFetch1[ORAS Fetch]
        ORASFetch2[ORAS Fetch]
        
    end

    %% Define relationships
    GetSubjectDescriptor -->|Store operation| GetSubjectReferenceManifest
    GetSubjectReferenceManifest -->|Store operation| Cache_Hit_1
    Cache_Hit_1{Cache Hit?}
    Cache_Hit_1 -->|Hit| GetReferrerArtifact
    Cache_Hit_1 -->|Miss| ORASFetch1
    ORASFetch1 -->|Fetch| MemoryBasedStore
    GetReferrerArtifact --> Cache_Hit_2
    Cache_Hit_2{Cache Hit?}
    Cache_Hit_2 -->|Hit| PerformVerification
    Cache_Hit_2 -->|Miss| ORASFetch2
    ORASFetch2 -->|Fetch| MemoryBasedStore
    PerformVerification --> ReturnVerifierResult
    OCIStoreCache --> CacheProvider
    CacheProvider --> OCIStoreCache

    %% Highlight subprocesses
    GetReferrerArtifact:::subProcess
    ORASFetch1:::subProcess
    ORASFetch2:::subProcess
    PerformVerification:::subProcess

    %% Add connections to the cache
    Cache_Hit_1 -.-> OCIStoreCacheWorker
    MemoryBasedStore -.-> GetReferrerArtifact
    Cache_Hit_2 -.-> OCIStoreCacheWorker
    MemoryBasedStore -.-> PerformVerification
    MemoryBasedStore -.-> OCIStoreCacheWorker
    OCIStoreCacheWorker -.-> CacheProvider
    

    %% Define external components
    classDef external fill:#f0f0f0,stroke:#333,stroke-width:2px;
    OCIStoreCacheWorker:::external
    MemoryBasedStore:::external
    OCIStoreCache:::external
    CacheProvider:::external
```

## Component Description

- Referrer Store Architecture

- Cache Provider: Set up to cache data for `GetSubjectDescriptor`, `GetReferenceManifest`.

- OCI Store(local cache): A content store based on file system with the OCI-Image layout.
  - Both executor and plugin initiation would `CreateStoresFromConfig`, add cache check can help avoid duplication.
  - task handler is event driven controlled by cache worker: new enqueue item created, last task finished
  - a write lock is set in Cache Provider to avoid conflict
  - a read lock/availability map is set for the writing content to avoid dirty read

- Cache Worker:
  - Manage memory backed content store
  - Enqueue task in cache provider to write into OCI Store
  - Check read lock/availability map in cache provider to avoid dirty read

```golang

type CacheWorker interface {

    // CreateMemoryOCIStore creates memory based OCI Store
    CreateMemoryOCIStore () error

    // GetAvailableOCIStore returns current OCI layout based OCI Store if exists
    GetAvailableOCIStore() (*orasStore, error)

    // GetBlobContent returns the blob with the given digest
    // WARNING: This API is intended to use for small objects like signatures, SBoMs
    GetBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) ([]byte, error)

    // GetReferenceManifest returns the reference artifact manifest as given by the descriptor
    GetReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) (ocispecs.ReferenceManifest, error)

    // GetCachedResource checks read lock from cache provider and returns data cached in OCI layout based OCIStore
    GetCachedResource() error

    // EnqueueCacheTask adds caching task into message queue in Cache Provider
    EnqueueCacheTask() error

    // DequeueCacheTask handle the up-coming caching task
    // checks write lock, availability map of resource
    DequeueCacheTask() error

    // CacheBlobContent cached target blob content into OCI layout based OCI Store
    CacheBlobContent(ctx context.Context, subjectReference common.Reference, digest digest.Digest) error

    // CacheReferenceManifest cached target manifest into OCI layout based OCI Store
    CacheReferenceManifest(ctx context.Context, subjectReference common.Reference, referenceDesc ocispecs.ReferenceDescriptor) error
}

```

## Comparing with Other Design

### Single OCIStore VS multi-OCIStores(per verifier)

1. Syncing in between those OCI stores will cause design difficulty.
2. If if OCI stores are verifer-binded, when the verifier is recycled the OCI stores is gone. Otherwise another resource maintainer(provider) is needed.
3. Maintaining multi-OCIStores means extra memory cost

### Async VS Sync

1. Sync write to OCIStore leads to racing condition. Using lock would naturally heads to more waiting in line.
2. When doing sync write, both read write racing condition have be handled in verifier process which increase the complexity. When doing async write, those jobs are delegated to **Cache Provider**.

## Supported Limits and Further Considerations

In Ristretto using scenario, do we support multi-notation verifier, in other words do we have to support cache worker with Ristretto

Message queue handling is event driven: when ever enqueue, dequeue finished cache worker should start trying dequeuing a new task

## Appendices

- [Multi-tenancy](https://ratify.dev/docs/reference/multi-tenancy)
- [Image Integrity](https://learn.microsoft.com/en-us/azure/aks/image-integrity?tabs=azure-cli#how-image-integrity-works)