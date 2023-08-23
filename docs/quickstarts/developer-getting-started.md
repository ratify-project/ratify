# Overview

The intent of this document is to be a high level quick start guide to get up and running quickly. For a more in depth review on configuration please see [CONTRIBUTING.MD](../../CONTRIBUTING.md#running-the-ratify-cli).

# Getting started

High-level steps of this document:

1. Start devcontainer
1. Set aliases and environment variables
1. Build images
1. Create Kind cluster
1. Install Gatekeeper (from repo via Helm)
1. Install Ratify (from locally built image)
1. Install a verifier
1. Apply constraints and templates
1. Validate

## Devcontainer

Start the devcontainer using the command pallet or the green button in the lower left corner of VSCode.

## Create aliases

```bash
alias k="kubectl"
```

## Export variables

```bash
export RATIFY_NAMESPACE=gatekeeper-system
export KUBERNETES_VERSION=1.25.4
export GATEKEEPER_VERSION=3.13.0
export IMAGE_PULL_POLICY=IfNotPresent
export RATIFY_LOG_LEVEL=INFO
```

## Build images

### Ratify

```bash
docker build \
    --progress=plain \
    --no-cache \
    -f ./httpserver/Dockerfile \
    -t localbuild:test .
```

### CRDs

```bash
docker build \
    --progress=plain \
    --no-cache \
    --build-arg KUBE_VERSION=${KUBERNETES_VERSION} \
    --build-arg TARGETOS="linux" \
    --build-arg TARGETARCH="amd64" \
    -f crd.Dockerfile \
    -t localbuildcrd:test ./charts/ratify/crds
```

## Kind

### Create cluster

```bash
kind create cluster
```

### Delete cluster

```bash
kind delete cluster
```

### Load images into kind

```bash
kind load docker-image --name kind localbuild:test
kind load docker-image --name kind localbuildcrd:test
```

## Gatekeeper

### Add repo

```bash
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
```

### Install

```bash
helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --version=${GATEKEEPER_VERSION} \
    --set validatingWebhookTimeoutSeconds=5 \
    --set mutatingWebhookTimeoutSeconds=2    
```

## Ratify

### Install

Install Ratify using TLS and a self signed cert.

Notes:

- See [other ways](../../CONTRIBUTING.md#deploy-from-local-helm-chart)  to install and TLS/mTLS options.
- If changes are made to a plugin, they will have to be re built using `make build-plugins` and copy the output to `./ratify/plugins/`.

1. Supply a certificate to use with Ratify (httpserver) or use the following script to create a self-signed certificate.

    ```bash
    ./scripts/generate-tls-certs.sh ${RATIFY_NAMESPACE}
    ```

1. Install ratify using a certificate

```bash
helm install ratify ./charts/ratify \
    --namespace ${RATIFY_NAMESPACE} --create-namespace \
    --atomic \
    --set provider.tls.skipVerify=false \
    --set provider.tls.cabundle="$(cat certs/ca.crt | base64 | tr -d '\n\r')" \
    --set provider.tls.key="$(cat certs/tls.key)" \
    --set provider.tls.crt="$(cat certs/tls.crt)" \
    --set image.repository=localbuild \
    --set image.crdRepository=localbuildcrd \
    --set image.tag=test \
    --set image.pullPolicy=${IMAGE_PULL_POLICY} \
    --set logLevel=info      
```

### Upgrade

```bash
helm upgrade -i ratify ./charts/ratify \
    --namespace ${RATIFY_NAMESPACE} --create-namespace \
    --atomic \
    --set provider.tls.skipVerify=false \
    --set provider.tls.cabundle="$(cat certs/ca.crt | base64 | tr -d '\n\r')" \
    --set provider.tls.key="$(cat certs/tls.key)" \
    --set provider.tls.crt="$(cat certs/tls.crt)" \
    --set image.repository=localbuild \
    --set image.crdRepository=localbuildcrd \
    --set image.tag=test \
    --set image.pullPolicy=${IMAGE_PULL_POLICY} \
    --set logLevel=info    
```

### Uninstall

```bash
helm uninstall ratify \
    -n ${RATIFY_NAMESPACE} \
    --debug
```

## Apply CRD

Install a the SBOM verifier via a CRD definition.

```bash
kubectl apply -f ./config/samples/config_v1alpha1_verifier_sbom.yaml
```

## Kubernetes constraints and templates

The constraint targets the 'default' namespace so any deployments to that namespace will be subject to this constraint. The sample constraint must be delete, and added back after, before any helm install or upgrading commands are run. If not, the deployment will fail as it will not not meet the constraints requirements.

### Install

```bash
kubectl apply -f ./library/default/template.yaml
kubectl apply -f ./library/default/samples/constraint.yaml
```

### Delete

```bash
kubectl delete -f ./library/default/samples/constraint.yaml
kubectl delete -f ./library/default/template.yaml
```

## Validate running pods

### All namepaces

```bash
kubectl get pods -A
```

### Ratify namespace

```bash
kubectl get pods -n ${RATIFY_NAMESPACE}
```

# K8s constraint validation

```bash
kubectl run demo --image=wabbitnetworks.azurecr.io/test/notary-image:signed
kubectl run demo --image=wabbitnetworks.azurecr.io/test/notary-image:unsigned
```

# Debugging

In VSCode hit F5, the cli will be called and a sample image will be verified.

See [debugging Ratify with VSCode](../../CONTRIBUTING.md#debugging-ratify-with-vs-code)

## Logs

### Ratify logs 

When installing Ratify the log level can be specified by specifying the switch `--set logLevel=info`.

The log level can also be configured by setting the env variable `RATIFY_LOG_LEVEL` with one of the follow values:

  - `PANIC`
  - `FATAL`
  - `ERROR`
  - `WARNING`
  - `INFO` (default)
  - `DEBUG`
  - `TRACE`

### Pod logs

use -p to see terminated pod logs, this is helpful when a pod starts and crashes.

```bash
kubectl logs <pod-name> \
    -n ${RATIFY_NAMESPACE} \
    --since=1h
```

### Deployment logs

```bash
kubectl logs deployment/ratify -n ${RATIFY_NAMESPACE}
```

## Clean up

```bash
kubectl delete -f ./library/default/samples/constraint.yaml
kubectl delete deployment ratify -n ${RATIFY_NAMESPACE}
kubectl delete po ratify-update-crds-hook-<foo> -n ${RATIFY_NAMESPACE}
kubectl delete po ratify-<foo> -n ${RATIFY_NAMESPACE}
```

## Useful commands

### Deployments

```bash
kubectl get deployments -A
kubectl describe deployment -n ${RATIFY_NAMESPACE}
kubectl rollout restart deployment ratify -n ${RATIFY_NAMESPACE}
```

### Configmaps

```bash
kubectl get configmap ratify-configuration -n ${RATIFY_NAMESPACE} -o json
kubectl edit configmap/ratify-configuration -n ${RATIFY_NAMESPACE}
  ```

### Pods

```bash
kubectl get pods -A
kubectl describe pod -n ${RATIFY_NAMESPACE}
```
