Ratify's configuration consist of [Store](store.md), [Verifier](verifier.md), [Policy](policy-provider.md) and [Executor](executor.md). 

## Configuration file
When Ratify runs in cli serve mode, configuration file can be dynamically updated while the server is running, subsequent verification will be based on the updated configuration.

## CRDs
Ratify also support configuration through K8 [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)    configuration can be updated using natively supported kubectl commands.

When running Ratify in a pod, configMap will be mounted to the pod at the default configuration file path. Ratify will initialize with specifications from the configuration file, CRDs will override store and verifier defined from the configuration file if they exist at runtime. Our team is in process converting configuration components into Ratify CRDS to support a more native k8s experience. Please see ratify CRDs samples in the [TODO charts](../charts/ratify/crds/) directory.

Currently supported components through CRDs are:

- Verifiers
- Stores
- Certificate Stores

### Get Crds
Our helms are charts are wired up to intialize CRs based on chart values. 
After Ratify installation, you can use the get command to review currently active configuration.

Sample command:

kubectl get stores.config.ratify.deislabs.io --namespace default
kubectl get verifiers.config.ratify.deislabs.io --namespace default
kubectl get certificatestores.config.ratify.deislabs.io --namespace default

### Update Crds
You can choose to add / remove / update crds. 
Here are some sample command to add or remove verifiers

toAdd:
kubectl apply -f .../ratify/config/samples/config_v1alpha1_verifier_cosign.yaml ( TODO switch to something else)
this enables the cosign verifier to Ratify

toRemove:
kubectl delete verifiers.config.ratify.deislabs.io/verifier-notary // delete a notary verifier
