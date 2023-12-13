# Cache Unification in Ratify
Author: Akash Singhal (@akashsinghal)

## Goals

- Define a generic Cache interface
- Implement default cache implementation using redis (distributed) and/or Ristretto (in memory)
- Remove all existing cache implementation
- Migrate existing cache implementation to use interface
    - This will require some refactoring of auth cache and subject descriptor cache
- Update caching documentation

## Non Goals
- Metrics support for new cache
- Persistent blob caching for ORAS OCI Store
- Certificate store cert caching

## Brief Aside about High Availability Goals for Ratify

Ratify is at risk as being a single point of failure in the admission process in K8s. Gatekeeper by default fails open so failures in reconciling the admission result of a particular resource will not block resource creation. However, this provides a weak security guarantee which is directly dependent on the availability for Ratify. For large clusters with constant resource creation, a single Ratify pod could be more susceptible to failure. The goal is for Ratify to be deployed as a replicaset. 

Many in-memory resources including all in-memory caching, file-based blob stores, and certificate files must be shared between the replicas to avoid very expensive network operations to the same resources. Remote resources like registries have throttling limits. Replicas that don't share common resources will almost certainly trigger throttling.

Ratify must:
1. Unify all of the memory based caches
2. Implement a distributed cache provider to be shared by multiple ratify processes and plugins
3. Add a persisent file storage for blob storage

## Overview

Please reference this [doc](https://ratify.dev/docs/reference/cache) for overview of current caching state in Ratify

Ratify has two primary cache categories: in memory caches & blob store cache.
There are 4 separate in-memory caches backed by 3 different cache types. This makes it very difficult to standardize cache interactions and emit uniform metrics. Furthermore, supporting multiple cache types will make it difficult to easily switch between in-memory and distributed caching for high availability scenarios.

Each of these levels of caching is integral to nearly every operation for ratify. All verification operations, all registry operations, and all auth-related operations rely on caching to provide performant execution.

During the design process, we discovered that the existing in memory caching strategy will not be suitable for external verifier/store plugins. Ratify invokes each external plugin as a separate process and as a result any shared in memory caches during the main ratify execution process cannot be shared with the sub plugin processes. <b>This particularly affects external verifier plugins which instantiate and interact directly with the ORAS store to perform registry operations as required.</b> Without a shared global cache with the main process, there could be potentially many redundant and expensive auth and descriptor requests to a remote registry.

<b>Should we use a distributed redis cache with a persistent OCI blob store?</b>

- Pros
    - External verifiers can share ORAS store cached info (Note: this is regardless of replicas or not)
        - Auth operations are time intensive and every invocation of an external verifier will trigger the auth flow
    - Blobs for manifest and layers will be downloaded once across all replicas
        - blob downloads are time and memory intensive. Each replica will have it's own duplicate copy of the blob if it's not shared
    - Duplicate requests to ratify could land on different replicas leading to extra operation overhead. Ratify's http server cache was created to avoid this but if it's not shared across replicas, then each replica will redo the first redundant operation it encounters.
- Cons
    - Redis will require either a pod singleton or a stateful set backed by a PVC which adds complexity and more dependencies. A pod singleton will itself be a potential single point of failure.
    - Blob store PVC will require a cloud-specific PV based on the CSI drivers available in cluster.
        - For Azure, we should use an Azure Blob CSI with azure-blob-nfs storage class. This is not enabled by default on the cluster and will require the user to manually enable on their AKS cluster. See [here](https://learn.microsoft.com/en-us/azure/aks/azure-csi-blob-storage-provision?tabs=mount-nfs%2Csecret) for more info.

<b>Should we change the verifier's responsibility so it does not interact with Referrer Store?</b>
Every verifier implementation would only have access to a byte array to validate against. It would be up to the executor to extract the blob content and invoke the verifier after.

- Pros
    - Would allow for ORAS store caching operations to be backed by cache
    - Would allow for easier metric emission
    - Avoid constant ORAS store recreaction for external verifiers
- Cons
    - Pretty big breaking change. All existing external and internal verifier would have to be rewritten
    - Some verifiers have multiple layers in a single artifact. And the composition of the results from each of those layers determines the overall result. For example, cosign will have multiple sigs as layers in the Image manifest but the overall verification result for cosign verifier is if at least one sig was found to be valid.


## Proposed Design

### Cache Interface
```
type Cache interface {
    // Get returns the value linked to key. Returns true/false for existence
    Get(ctx context.Context, key string) (interface{}, bool)
    
    // Set adds value based on key to cache. Assume there will be no ttl
    Set(ctx context.Context, key string, value interface{}) bool
    
    // SetWithTTL adds value base on key to cache. Ties ttl of entry to ttl provided
    SetWithTTL(ctx context.Context, string, value interface{}, ttl time.Duration) bool
    
    // Delete removes the specified key/value from the cache
    Delete(ctx context.Context, key string) bool
    
    // Clear removes all key/value pairs from the cache (mainly helpful for testing)
    Clear(ctx context.Context) bool
}
```

### Cache Factory Interface
```
type CacheFactory interface {
    // Create instantiates a cache provider implementation. It returns the prvoider and returns if creation was successful.
    Create(ctx context.Context, cacheEndpoint string) (CacheProvider, bool)
}
```

### Cache Implementation

Each cache implementation will follow the factory/provider paradigm of other ratify components: A cache implementation will register a factory `Create` method with the global factory provider map. The cache-specific factory will be responsible for creating an instance of the cache and returning it. The `NewCacheProvider` function will be responsible for calling the corresponding `Create` function for the cache implementation name specified. Finally, the static `memoryCache` variable will be set to the created cache instance. An accessor method will be used throughout ratify packages to retrieve the global cache reference.

The intialization of the cache will be done only for the `serve` CLI command. New flags for `cache-type` and `cache-endpoint` can be specified to override the default values of ('redis', and 'localhost:6379'). Cache initialization is a blocking operation that will stall Ratify startup if not successful.

The cache interface returns whether each operation was successful or not rather than the error. All cache operations should be non-blocking and will fall-back to a non cache implementation if any errors occur. It is left to the cache implementation to log any errors so users are informed.

A well known cache key schema must be defined for each caching purpose. This will avoid Key conflicts. 
```
"cache_ratify_subject_descriptor_{SUBJECT_REFERENCE_DIGEST}"
"cache_ratify_list_referrers_{SUBJECT_REFERENCE_STRING}" // not guaranteed to be digest
"cache_ratify_verify_handler_{SUBJECT_REFERENCE_STRING}"
"cache_ratify_oras_auth_{SUBJECT_REGISTRY_HOST_NAME}"
```

### Redis Cache Implementation

The `Create` method will initialize a new `go-redis` client with with the provided `cacheEndpoint`. Each of the interface methods will implement the corresponding Redis accessor, writer, deleter, and clear methods.

Redis is key/value string typed. Any values passed will be JSON marshaled into string form before being written to cache. This requires all value types to be marshalable. Similarly, all values returned from the cache will be generic string typed. Since the redis cache implementation is agnostic of the value passed in, it is left to the invoker to unmarshal the `interface{}` into the desired type.

### Ratify with Redis Setup

- Http Server (non K8s)
    - By default the cache will be optional and not enabled. If the user decides to enable it, they are responsible for standing up a redis instance and providing the correct cache endpoint
- K8s: Redis will be a part of the ratify helm chart. A ClusterIP Service will expose redis inside the cluster at a well known endpoint.

### Redis security concerns

Please refer to [this document](https://hackmd.io/@akashsinghal/S1tcdavvh) for detailed Redis security analysis

## Open Questions
1. Is it worth defaulting to redis or should we just focus on in memory cache behind ristretto? 
    - The tradeoffs are listed in the Overview section
    - [UPDATE]: We have decided to default to ristretto and move redis cache behind an feature fla
3. What type of Redis installation should ratify support in helm chart?
    - Redis Pod singleton: Stand up a single Redis Pod shared between all replicas. Simpler but will not be persistent beyond pod lifespan. Lower availability since it's singleton.
    - Redis Cluster: A stateful set which will be more robust but has more dependencies. Adds complexity and may be unnecessary at this point
    - [UPDATE]: We will ship with a basic Redis Pod. NOTE: It's up to user to configure a production ready redis instance
4. Which Redis image should we use? GHCR? Dockerhub? MCR?
    - [UPDATE]: redis image from docker hub pinned to specific digest
6. How do we handle memory requirements for Redis pod? 
    - [UPDATE]: we will ship with a single redis pod with default 256 Mb requested
8. Show we also support Ristretto for an in-memory solution if a user is only going to use built in verifier/store? (ORAS + notation)
    - [UPDATE]: we will ship with both ristretto and redis

<hr/>

## ADDENDUM 6/21/23 Dapr Exploration

[Dapr](https://docs.dapr.io/) (Distributed Application Runtime) is a portable runtime that allows applications to easily integrate with many resources and services. It's built for distributed applications which adopt a micro-service architecture. The centralized control of the various integration points allows users to build platform agnostics microservices.

Dapr in K8s is deployed as a sidecar container on each pod. It intercepts all requests made through the Dapr client. The Dapr Operator Pod is reponsible for managing the various Dapr Integration Resource CRDS generated. Each CR represents a 3rd party supported integration. Dapr Sentry Pod is responsible for injecting the trusted certs used by the side car for all requests (mTLS). The Dapr Side-Car Injector watches for new pod creations and injects Daprd side containers as necessary. 

Dapr has robust [state-store support](https://docs.dapr.io/developing-applications/building-blocks/state-management/state-management-overview/). This shim allows the application to be state-store agnostic. There's support for many many different state stores including Redis. The application initiates a Dapr client using the sdk. A state-store specific resource is applied on the cluster to point configure Dapr. 

![](https://hackmd.io/_uploads/H13xF1bOh.png)
![](https://hackmd.io/_uploads/BkinY79_3.png)



Pros:
- Resiliency
    - Automatic detection of request failures to state store
    - built in retry policies with customizable circuit breakers
    - self heals if problem mitigates on its own
- Secure communication
    - mTLS communication with State store by default
    - automatic cert creation/rotation
- Data encryption at rest
    - AES Cipher with Galois Counter Mode
    - Automatic encryption/Decryption of all cache values for set/get ops
    - This is critical for securely caching registry credentials. Please see [this document](https://hackmd.io/@akashsinghal/S1tcdavvh) for more details.
- Health Reporting
    - Metrics are reported for Dapr control plane and sidecar components
    - Integration with Prometheus and Grafana
- Works with any state store
    - Ratify will recommend Redis but any state store can be used including cloud provider specific ones
- Works in self hosted scenarios
    - Ratify has maintained feature parity with running Ratify as a standalone HTTP Server
    - Dapr has a non-K8s self-hosted mode which could be used in stand-alone Ratify scenarios
- Large open source community with great documentation and support

Cons:
- Dapr must be installed on cluster
    - New prereq for ratify to work in replica set mode
    - Each pod will have a side car container which has increase resource consumption
    - Gatekeeper's pub/sub integration also requires Dapr/Redis to be preinstalled
- Multi-Tenancy is not supported out of the box
    - This is not a Dapr exclusive problem
    - Multi tenancy design will need to consider using an isolated Dapr instance per tenant

### Proposal

Cache unification design will remain: all 4 in memory caches will be unified behind a single `CacheProvider` implementation. The were will only be two cache providers:
    1. Ristretto: An in-memory cache provider. This will provide feature parity with Ratify's current capabilitiy. This will be the default mode used and should ONLY be used for Ratify single pod usage.
    2. Dapr: distributed state store shim. Equivalent Dapr sdk methods will be used in the `CacheProvider` interface implementation. NOTE: Installing Dapr and Redis will be documented pre requisities but the Ratify chart will NOT handle automatic installation (Gatekeeper is taking the same approach). Dapr will live behind a feature flag and be turned off by default.

<hr />

## ADDENDUM 5/24/23 Sharing Redis with GK

Gatekeeper is introducing support for publishing constraint violations to external sources via pub-sub providers. The first provider they have added is Dapr. The message broker selected with Dapr is Redis. Ratify's should plan to share the same Redis instance as GK's pub-sub integration. 

### Follow-up questions
1. What is redis perf/mem consumption at 2 million pub-sub message plus Ratify in 10k pod cluster?


### Action Items
- Conduct perf/mem testing with GK when redis cache provider is merged to main in mid June
<hr />

## Following portion of the doc is out of date and will not be updated

Further investigation revealed that there's a hard requirement for caches to be shared across processes. External plugins for verifiers/referrer stores are invoked as separate processes by Ratify. This means an in memory cache cannot be accessed by external plugin processes.

Assumptions:
- There are 3 scenarios that require caches: ratify as a replica set in K8s, ratify as a single pod in K8s, ratify as an independent http server
- CLI commands don't need a cache. 
    - If there every becomes a scenario where a single verification via `verify` command on Ratify CLI requires caching, user can point ratify to a preconfigured Redis instance like the `serve` command

## Proposal

Ratify does not enable a cache by default. All cache references to a unified cache would check to make sure cache is initialized or not. If not initialized it will skip any cache interaction and treat it as a cache miss. 

- Http Server
    - The user will be responsible for standing up a redis instance, specifying the Redis cache type and Redis port/IP as flags to the `serve` command. (Note: we could potentially explore having the `serve` command start up a docker container with redis on it but don't think this is ideal). 
- K8s: We could treat the single pod Ratify to be a base case of the replica set case. This would mean there's always a single Ratify Redis pod deployed regardless of replica count. Note: Cache interactions will be slower than using localhost. 
- K8s with side car container: If we want to special case the Ratify single pod scenario, we could have redis deployed as a separate container on the same Ratify pod. (Personally not in favor of this special casing)



