# Caching in Ratify

Ratify supports many levels of in-memory and file-store-based caching in order to reduce operation overhead and lower the overall verification and mutation duration.

## In Memory Caches

There are four different in-memory caches used for separate purposes. There are three different cache implementations as well.

### K8s API Machinery LRU Util Cache

The util cache, implemented in k8s apimachinery package, is an LRU in memory cache. See [docs](https://pkg.go.dev/k8s.io/apimachinery/pkg/util/cache#LRUExpireCache). It is used as the http server verification handler cache. This cache acts as a short lived response cache. In particular, this cache deduplicates redundant cache requests received by Ratify at once.

### Sync Map

This is a thread-safe mutex-lock map defined in the default go `sync` package. This is a very rudimentary data store that we use as a cache. See [docs](https://pkg.go.dev/sync#Map).

The authentication cache for ORAS is backed by a sync map. This cache is a key value mapping between the fully qualified subject reference string (hostname, repository, and tag) and an ORAS repository client initialized with credentials fetched from the Authentication Provider specified. 

The subject descriptor cache for ORAS also uses a sync map. It maps a digest based full qualifed subject image reference to a resolved subject descriptor. This is used by the `GetSubjectDescriptor` ORAS store implementation. The `/mutate` path heavily relies on this cache due to the often redundant mutate requests Ratify gets during K8s resource creation.

### Ristretto

Ristretto is a highly performant caching library. Currently, Ratify uses a Ristretto cache for a cached ORAS store implementation. There is only support for `ListReferrers` cached-based implementation. The cache implementation of `ListReferrers` stores the list of artifact descriptors returned from the registry referrers call. The ristretto cache is configured as LRU with a short TTL that can be configured via Helm. 

See [docs](https://github.com/dgraph-io/ristretto).

## File store based cache

ORAS provides an OCI layout store for caching blobs in a local file descriptor. Ratify's ORAS store implementation stores all blobs fetched from registry in an OCI store. During verification, blob-related operations (`GetReferenceManifest` & `GetBlobContent`) check the OCI file store for blob existence before making calls to registry.

## Flow Diagrams

![](https://i.imgur.com/VGIb3lu.png)
![](https://i.imgur.com/71l0jMG.png)
![](https://i.imgur.com/8w0OK6Q.png)

