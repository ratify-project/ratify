# Ratify Concurrency Implementation
Author: Akash Singhal (@akashsinghal)

Currently, Ratify is not concurrent. It cannot verify multiple subjects, multiple stores, or multiple artifacts in parallel. This will lead to serious performance degradation for complicated scenarios eventually hitting the validating webhook timeout. Ratify will begin by implementing go routines at the subject, store, and artifact level. Furthermore, some update and cache operations will be refactored to use locks.

## Adding Go Routines

- Multiple Subjects
    - The http server handler is responsible for invoking the executor's verify function for each subject reference provided in the External Data request. This inner loop will be converted to it's own go routine.
- Multiple Stores / Multiple Artifacts
    - The Executors core execution loop iterates through all configured stores, extracts all artifacts for the subject, and then runs the corresponding verifier for each artifact. These nested operations will be changed to a series of nested go routines (go routines that spawn other go routines). 

## Adding Locks
### ORAS Authentication Cache
- Switch auth cache from regular cache to `sync.Map` to guarantee atomic write updates

## Addendum 10/6/22

After initial implementation, we see significant performance improvements.

### Testing Scenario
Deployment with 4 containers each with 5 notary signatures attached. This means ratify will have to verify 20 signatures.

After multiple runs, the entire ratify execution is taking between 2.7 - 3.3 seconds to complete.

Sample Deployment YAML:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app: test-deployment
spec:
  replicas: 1 # testing purposes only
  selector:
    matchLabels:
      app: test-deployment
  template:
    metadata:
      labels:
        app: test-deployment
    spec:
      containers:
      - name: nginx1
        image: artifactstest.azurecr.io/notary/subjects/nginx:signed
      - name: nginx2
        image: generaltest.azurecr.io/notary/subjects/nginx:signed
      - name: alpine1
        image: artifactstest.azurecr.io/notary/subjects/alpine:signed
      - name: alpine2
        image: generaltest.azurecr.io/notary/subjects/alpine:signed
        
```

### ORAS throttling

