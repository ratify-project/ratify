Ratify's configuration consist of [Store](store.md), [Verifier](verifier.md), [Policy](policy-provider.md) and [Executor](executor.md). 

When Ratify runs in cli serve mode, configuration file can be dynamically updated while the server is running, subsequent verification will be based on the updated configuration.

When running Ratify in a pod, configMap will be mounted to the pod at the default configuration file path. Ratify will initialize with specifications from the configuration file, CRDs will override store and verifier defined from the configuration file if they exist at runtime. Our team is in process converting configuration components into Ratify CRDS to support a more native k8s experience. Please review ratify CRDs in the [charts](../charts/ratify/crds/) directory.