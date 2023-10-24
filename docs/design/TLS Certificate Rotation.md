# TLS Certificate Rotation in Ratify
Author: Akash Singhal (@akashsinghal)

Ratify supports TLS and mTLS communication with Gatekeeper when handling external data provider requests. Certificates and keys managed by Gatekeeper/Ratify can be rotated independently.

The Ratify server should be able to:
1. Identify changes to any of these certificates
2. Replace TLS certs with new versions
3. Reduce request dropping due to TLS cert mismatch (ideally none)

## How it works currently

mTLS guarantees communication integrity from both sides. Client and server must verify the other parties respective certificates. Gatekeeper generates it's own TLS public certificate and key which are derived from a (new or existing) CA certificate and key. Ratify is provided the CA public key during startup. 

Similarly, Gatekeeper needs access to Ratify's CA public key to perform verification. This is provided via Gatekeeper's `Provider` CRD. On startup, Ratify takes in a CA cert/key and tls cert/key. If not provided in the helm chart, this bundle is automatically generated during Ratify installation. The TLS certificate and key are configured in Golang's `http.server` TLS config for all responses sent back to Gatekeeper. 

Gatekeeper's CA cert can be rotated at any time AND Ratify's CA along with TLS cert/key can also be rotated. Currently, these 3 files are loaded once from disk on Ratify startup. Thus, any changes to these files, even if they are updated on disk, will not be picked up by Ratify's server instance without a <b>manual</b> pod restart. A pod restart will also result in potential requests being dropped and cache being reset. 

Gatekeeper's CA public key is stored as a K8s `Secret`. GK enables a certificate controller that automatically rotates certificates and updates this secret. Ratify mounts the secret as a volume in the pod, thus changes to the secret will get updated in the container file system. Ratify's CA cert/key & TLS cert/key are also stored as a K8s secret and mounted in a similar fashion. 

There is an unfortunate side effect to relying on K8s secret volume mount. K8s secret in general have a 60-90 second delay in updating the volume mount. See [this article](https://ahmet.im/blog/kubernetes-secret-volumes-delay/) for more detailed explanation. This means that even if Ratify always uses the current certificate mounted in the container, the GK CA cert may be out of date for the sync duration. In that time, any request will result in TLS error from Ratify. Other projects such as Eraser project have circumvented this issue by forcing an immediate trigger of the pod sync operation. This is achieved by a controller applying a dummy annotation to the Ratify pod. See [here](https://github.com/Azure/eraser/blob/5b96b5cfabb95671db7aff588b73662fbfcdacbc/controllers/configmap/configmap.go#L155-L164)

## Proposed Solution

~~- Implement a custom TLS config reload function for `http.server`
    - Each new TLS connection will call this function first
    - Function will reload 3 certs from disk every time. This will increase the file read operations. Not sure if this will cause issues in the future for very large scale scenarios
    - Pros: 
        - Relatively simple to implement
        - Will avoid manual restart of pod currently required
    - Cons: 
        - Expensive read ops which potentially could pose a problem in the future~~

- Implement a custom TLS config reload function for `http.server` and use a file watcher client and server cert file paths
    - Similar to the solution above but relies on a file watcher to watch for updates instead of reading from disk every time
    - Base implementation off of K8s cert watcher [here](https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/certwatcher/certwatcher.go) but also need to support clientCA update.
    - Pros:
        - Avoid manual pod restart
        - Avoid cert file reading for each TLS connection
    - Cons:
        - more complex: will require implementing custom watcher and update functionality (potentially more prone to error)
- If we need to address the 60-90 second delay, we need to implement a controller to watch the secrets.
    - Controller will update the ratify pods with a dummy annotation if GK or ratify secret changes
    - Pro:
        - Will reduce chance of request dropping due to cert rotation to almost 0
    - Cons:
        - Will take longer to implement and more involved 
        - Is this a good design pattern to own a controller for a resource not managed by Ratify?

I believe we should start with updating the Ratify server. This will mitigate the immediate issue. Follow up work can be done to implement a new controller.


## Open Questions
1. Is an initial 60-90 second downtime manageable initially?

## How to test this is working correctly?

- Generate a new cert bundle for GK and manually update GK webhook cert. Wait 2 minutes and then check that subsequent requests to ratify are still successful. 
- Generate a new cert bundle for Ratify. Update GK `Provider` CR with new CA bundle. Wait 2 minutes and check that subsequent requests to ratify are still successful.