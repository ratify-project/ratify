Ratify's configuration consist of [Store](store.md), [Verifier](verifier.md), [Policy](providers.md#policy-providers) and [Executor](executor.md). 

## Configuration file
When Ratify runs in cli serve mode, configuration file can be dynamically updated while the server is running, subsequent verification will be based on the updated configuration.

## CRDs
Ratify also supports configuration through K8 [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). The configuration can be updated using natively supported `kubectl` commands.

When running Ratify in a pod, the `ConfigMap` will be mounted in the pod at the default configuration file path. Ratify will initialize with specifications from the configuration file. CRDs will override store and verifier defined in the configuration file if they exist at runtime. Our team is in the process of converting configuration components into Ratify CRDs to support a more native k8s experience. Please review ratify CRDs samples [here](../../config/samples/).

Currently supported components through CRDs are:

- [Verifiers](../reference/crds/verifiers.md)
- [Stores](../reference/crds/stores.md)
- [Certificate Stores](../reference/crds/certificate-stores.md)
- [Policy](../reference/crds/policies.md)

### Get Crds
Our helms charts are wired up to initialize CRs based on chart values. 
After Ratify installation, you can use the `kubectl` command to review the currently active configuration.

Sample command:
```bash
kubectl get stores.config.ratify.deislabs.io --namespace default
kubectl get verifiers.config.ratify.deislabs.io --namespace default
kubectl get certificatestores.config.ratify.deislabs.io --namespace default
kubectl get policies.config.ratify.deislabs.io --namespace default
```
### Update Crds
You can choose to add / remove / update crds. 
Sample command to update a verifier:
```bash
kubectl apply -f .../ratify/config/samples/config_v1alpha1_verifier_schemavalidator.yaml
```
Sample command to remove a verifier:
```bash
kubectl delete verifiers.config.ratify.deislabs.io/verifier-notation 
```