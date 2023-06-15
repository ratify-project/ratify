A `Store` resource defines how to discover and retrieve reference types for a subject.
Please review doc [here](https://github.com/deislabs/ratify/blob/main/docs/developer/store.md) for a full list of store capabilities. 
To see more sample store configuration, click [here](../../../config/samples/). Each resource must specify the `name` of the store.
 
```yml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Store
metadata:
  name: 
spec:
  name: required, name of the store
  address: optional. Plugin path, defaults to value of env "RATIFY_CONFIG" or "~/.ratify/plugins"
  source:  optional. Source location to download the plugin binary, learn more at docs/reference/dynamic-plugins.md
  parameters: optional. Parameters specific to this store
```

## Oras

An implementation of the `Referrer Store` using the ORAS Library to interact with OCI compliant registries.

Sample Oras yaml spec:
```yml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Store
metadata:
  name: store-oras
spec:
  name: oras
  parameters: 
    cacheEnabled: true
    capacity: 100
    keyNumber: 10000
    ttl: 10
    useHttp: true  
    authProvider:
      name: k8Secrets
      secrets: 
      - secretName: ratify-dockerconfig
```


| Name        | Required | Description | Default Value |
| ----------- | -------- | ----------- | ------------- | 
| cosignEnabled      | no    |   This must be `true` if cosign verifier is enabled. Read more about cosign verifier [here](https://github.com/deislabs/ratify/blob/main/plugins/verifier/cosign/README.md).        |   `false`       |
| authProvider      | no    |      This is only required if pulling from a private repository. For all supported auth mode, please review [oras-auth-provider](https://github.com/deislabs/ratify/blob/main/docs/reference/oras-auth-provider.md) doc  |   dockerAuth            |
| cacheEnabled      | no    |   Oras cache, cache for all referrers for a subject. Note: global cache must be enabled first     |   `false`            |
| ttl      | no    |    Time to live for entries in oras cache        |   10 seconds            |
| useHttp      | no    |  This needs to be set to `true` for  local insecure registries           |  `false`     |

