# OCI Store Index Race Condition

## Purpose
Resolve OCI Store race condition in multi-verifier scenarios

## Table of Content
- Data Flow
- Component Description
- User Scenarios
- Performance
- Supported Limits
- Appendices

### Data Flow
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

### Component Description
- Cache Provider: Set up to cache data for `GetSubjectDescriptor`, `GetReferenceManifest`.

- OCI Store(local cache): A content store based on file system with the OCI-Image layout.
  - Both executor and plugin initiation would `CreateStoresFromConfig`, add cache check can help avoid duplication.

- Cache Worker: 
  - Manage memory backed content store
  - Enqueue task in cache provider to write into OCI Store
  - Check read lock in cache provider to avoid dirty read

### User Scenarios
1. Using different version of same verifier

2. Multi-tenancy

### Performance

### Supported Limits

### Appendices
