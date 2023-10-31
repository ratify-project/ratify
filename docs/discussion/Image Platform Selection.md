# Ratify Image Platform Selection
Author: Akash Singhal (@akashsinghal)

## Problem

Currently, Ratify is not aware of the host OS/Architecture that a container should be used with. As a result, any subject reference that is an OCI Index is validated at the Index level instead of the platform-specific manifest that the container will use. This leads to two main questions:

1. How should Ratify treat an OCI Index?
    - Should the OCI Index be kept as an ordinary artifact with signatures verified ONLY at the index level?
    - Should the OCI Index be further resolved to the platform-specific manifest and the validation occurrs ONLY on the platform-specific manifest?
    - Should the OCI Index be validated AND it's platform-specific manifest be validated? 
2. How does Ratify retrieve the target host OS/Architecture of the container?
    - The OS/Arch that ratify runs on may not be the target os/arch of the container if there are mixed nodepools
    - The Kube scheduler determines which node the pod is assigned to. However, at admission time, the kubescheduler has not been called thus we have no node assigned at that point to extract OS/Arch.

## Proposed Solution

Ratify MUST special case the OCI Index and Docker Manifest List. The subject reference should resolve to the descriptor of the index/list and the platform-specific manifest in the index/list. We will change the behavior of the `GetSubjectDescriptor` method to return a list of descriptors which will all be treated as subjects to be verified. We will leverage the existing ORAS store `Resolve` function to provide a platform specific hint that can be used to return the correct descriptor. 

At admission time of the Pod resource, the `AdmissionRequest` resource that Gatekeeper receives as input, contains the full PodSpec. Since the Pod is yet to be created, the kube scheduler has not matched the pod to a node to run on in the cluster. Thus we can ONLY rely upon the information in the PodSpec. During policy evaluation, the rego has access to the entire PodSpec `input`. This spec can be used to find specific node selection labels such as `kubernetes.io/arch` and `kubernetes.io/os`. If found, we can provide the target OS/arch combination in the external data provider request to Ratify. Since the current data Provider request only accepts a list of strings, we propose to define a new string based ratify request that will encode the subject reference and the target OS/Arch information. Ratify will be able to interpret this string request and provide the OS/Arch hint to the referrer store's `GetSubjectDescriptor` method. If the external data provider request does not contain the OS/Arch for reference, Ratify will default to inferring the OS/Arch based on the node the Ratify pod is running on. 

Typically, user can specify the node os/arch a pod must run on using Node Selectors, Affinity, and Tolerations. The rego templates we provide should handle each scenario.

#### Node Selectors

This is the most common way to define OS/Arch in the pod spec. 

Example: 

```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  nodeSelector:
    kubernetes.io/arch: arm64
    kubernetes.io/os: linux
    
```

#### Affinity

Example:
```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
            - key: "kubernetes.io/arch"
              operator: In
              values: ["amd64"]
```
#### Taints and Tolerations

Example: Set a taint on the Nodes with `key: kubernetes.io/arch` and `NoSchedule` effect for "arm64".
```
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  tolerations:
  - key: "kubernetes.io/arch"
    operator: "Equal"
    value: "arm64"
    effect: "NoSchedule"
```

### New External Data Ratify Contract

```
{
    "reference": "ratify.azurecr.io/testimage:signed",
    "os": "linux",
    "arch": "arm64"
}
```

Payload would be JSON serialized and converted to string. Each entry in the list of parameters in external data provider request would a serialized string.
### Concerns

- Can we depend on the PodSpec to define the OS/Arch?
    - Are there instances where only one of the OS/Arch is defined? If so how do we handle this scenario?
- Are the built-in `kubernetes.io/arch` and `kubernetes.io/os` labels enough to rely upon? What happens when user has custom label names instead?
    - Possible solution: Sample rego template will enclude rego functions to extract OS/Arch ONLY using these labels. User will have to adapt rego if using custom labels. 
- NodeSelector is quite straightforward to write rego. Affinity is much more flexible which means there are multiple ways to define the `nodeSelectorTerms`.
    - How would we write flexible rego for Affinity and Tolerations?
        - Possible Solution: Offer a template function for Affinity and Toleration that handles the common scenarios provided above for each. If user has a different Affinity/Toleration, they will be required to alter it.

