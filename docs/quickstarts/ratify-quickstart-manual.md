
# Manual Quick Start Steps

This document outlines the **manual** production ready steps to install Ratify with Gatekeeper in admission control scenarios. Please refer to the [README.MD](../../README.md) for recommended install steps.

<hr>

Prerequisites:
- Kubernetes v1.20 or higher
- OPA Gatekeeper v3.10 or higher  

### Step 1: Setup Gatekeeper with [external data](https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata)

> NOTE: If you have added Helm repository for Gatekeeper and Ratify, you can update them by executing `helm repo update` before installation.

```bash
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=5 \
    --set mutatingWebhookTimeoutSeconds=2
```

> NOTE: `validatingWebhookTimeoutSeconds` and `mutationWebhookTimeoutSeconds` increased from 3 to 5 and 1 to 2 respectively, so all Ratify operations complete in complex scenarios. See [discussion here](https://github.com/deislabs/ratify/issues/269) to remove this requirement. Kubernetes v1.20 or higher is REQUIRED to increase timeout. Timeout is configurable in helm chart under `provider.timeout` section.   

### Step 2: Deploy ratify on gatekeeper in the gatekeeper-system namespace.

- Option 1: Install the last released version of Ratify

Note: if the crt/key/cabundle are NOT set under `provider.tls` in values.yaml, helm would generate a CA certificate and server key/certificate for you.

```bash
helm repo add ratify https://deislabs.github.io/ratify
# download the notary verification certificate
curl -sSLO https://raw.githubusercontent.com/deislabs/ratify/main/test/testdata/notation.crt
helm install ratify \
    ratify/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notaryCert=./notation.crt \
    --set featureFlags.RATIFY_CERT_ROTATION=true
```

- Option 2: Install ratify with charts from your local branch.  
Note: Latest chart in main may not be compatible with the last released version of ratify image, learn more about weekly dev builds [here](RELEASES.md/#weekly-dev-release) 
```bash
git clone https://github.com/deislabs/ratify.git
cd ratify
helm install ratify \
    ./charts/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notaryCert=./test/testdata/notation.crt \
    --set featureFlags.RATIFY_CERT_ROTATION=true
```

### Step 3: See Ratify in action

- Deploy a `demo` constraint.
```
kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
```

Once the installation is completed, you can test the deployment of an image that is signed using Notary V2 solution.

- This will successfully create the pod `demo`

```bash
kubectl run demo --namespace default --image=ghcr.io/deislabs/ratify/notary-image:signed
kubectl get pods demo
```

Optionally you can see the output of the pod logs via: `kubectl logs demo`

- Now deploy an unsigned image

```bash
kubectl run demo1 --namespace default --image=ghcr.io/deislabs/ratify/notary-image:unsigned
```

You will see a deny message from Gatekeeper denying the request to create it as the image doesn't have any signatures.

```bash
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: wabbitnetworks.azurecr.io/test/net-monitor:unsigned
```

You just validated the container images in your k8s cluster!

### Step 4: Uninstall Ratify
Notes: Helm does NOT support upgrading CRDs, so uninstalling Ratify will require you to delete the CRDs manually. Otherwise, you might fail to install CRDs of newer versions when installing Ratify.
```bash
kubectl delete -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl delete -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
helm delete ratify --namespace gatekeeper-system
kubectl delete crd stores.config.ratify.deislabs.io verifiers.config.ratify.deislabs.io certificatestores.config.ratify.deislabs.io policies.config.ratify.deislabs.io
```