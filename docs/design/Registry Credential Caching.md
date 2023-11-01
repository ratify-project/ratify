# Registry Credential Caching
Author: Akash Singhal (@akashsinghal)

Ratify stores the auth credentials for a registry in an in-memory map as part of the ORAS store. We want to unify the caches across Ratify so that we can introduce new cache providers such as Redis to support HA scenarios.

> Is storing registry credentials in a centralized cache a security risk?

During investigation of cache unification, we found that the external plugins cannot share in-memory caches with the main Ratify process since they are invoked as separate processes. This includes registry credentials. If we don't address this problem, external plugins will invoke authentication flow every time they are invoked which will be a huge performance bottleneck. 

> How do we make auth credentials accessible to external plugins?


## What is the security boundary?
Ratify will be deployed inside a user K8s cluster. Only users with correct RBAC can access the cluster but there's no guarantee how a centralized cache instance will be secured. For Redis, Ratify will store the registry credentials in a well-known cache key pattern which will allow any application with an authorized client (via a password) to read registry credentials. Redis has a concept of "Databases" but in reality they are only logical "namespace" separators and do not inherently provide data isolation. We cannot rely on "Database" partitioning to provide security guarantees. (See [this](https://github.com/redis/redis/issues/8099#issuecomment-741868975) discussion) Other applicatons are within the cluster boundary but can we assume those application fall into the trusted boundary?

Azure Redis Cache multitenancy [guidance](https://learn.microsoft.com/en-us/azure/architecture/guide/multitenant/service/cache-redis#isolation-models) shows that the only way to guarantee high data isolation, multiple cache instances must be used. 

![](https://hackmd.io/_uploads/BJKr42gd2.png)


## Potential Solutions

### 1. Do not use a central cache for registry credentials

* Each Ratify instance (pod) would continue to use an in-memory cache map for ORAS registry credential cache
* This cache would be separate from the other unified cache implementation.
- The cache lifespan would be tied to lifecyle of the process that creates it

- Pros:
    - This will provide the highest security guarantee as credentials do not leave the ratify container
    - Future auth providers may have dependence on registry credentials originating from the pod that fetched it. Other pods wouldn't be able to share these credentials in this case. There is no known scenario yet that requires this.
- Cons:
    - No support for external plugins
        - External plugins are separate processes and thus cache will be recreated in the process context.
        - Built-in verifiers like Notation will not be affected but all others will be
    - Dual cache strategy will complicate management and metric emission
    - Performance hit in replica set scenario (this might be minor and isolated to first auth flow for a new registry per pod)

### 2. Encrypt the credential cache values in Redis

* Implement data encryption functions for registry credentials
- Use AES Cipher in Galois/Counter Mode for inflight encrypt/decrypt for cache read/write. (Symmetric encryption)
* Store encryption key in a K8s secret

- Pros:
    - Centralized cache model
    - Credentials can be shared between pods
    - Security guarantee regardless of security of the centralized cache instance
- Cons:
    - Performance overhead for encryption/decryption operations
    - Key Rotation with cache invalidation?

- NOTE: Dapr provides built in encryption and key rotation capabilities out of the box that will work with Redis

### 3. Leverage K8s secrets and associated RBAC to enforce access policies

* Each unique registry host will have a K8s secret
* RBAC will be set so only Ratify can read from the secrets.
* K8s secret pusher and fetcher functions will handle equivalent cache push and evict operations

- Pros:
    - Relies on native K8s RBAC to provide security boundary
- Cons: 
    - K8s specific implementation
        - We have maintained relative parity with running Ratify in non-k8s. How do we handle non K8s caching of the creds?
    - Limited perf hit due to fetching resources using API server
    - Dual cache strategy will complicate management and metric emission

