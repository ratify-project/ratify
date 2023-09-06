# Install Ratify for High Availability

The default Ratify installation relies on a single Ratify pod processing all requests. For higher performance and availability requirements, Ratify can be set to run with multiple replicas and a shared state store.

Ratify installation/upgrade for HA scenarios can be done via a `helmfile` or manual installation steps. Both options are outlined in this document.

## Automated Installation
> Note: Helmfile does not have a stable release and thus is NOT recommended for production environments 
**Prerequisites**
```bash
# Download and install yq
	curl -L https://github.com/mikefarah/yq/releases/download/v4.34.2/yq_linux_amd64 --output /usr/bin/yq && chmod +x /usr/bin/yq
```
```bash
# Download and install helmfile
  curl -LO https://github.com/helmfile/helmfile/releases/download/v0.155.0/helmfile_0.155.0_linux_amd64.tar.gz
  mkdir helmfile-install
  tar -zxf helmfile*.tar.gz -C helmfile-install/
  mv helmfile-install/helmfile /usr/bin
  rm -rf helmfile*.tar.gz helmfile-install/
```

```bash
# Sync helm chart resources defined with cluster
curl -L https://raw.githubusercontent.com/deislabs/ratify/main/high-availability.helmfile.yaml | helmfile sync -f -
```

### Uninstall Steps
```bash
curl -L https://raw.githubusercontent.com/deislabs/ratify/main/high-availability.helmfile.yaml | helmfile destroy --skip-charts -f -
```
## Manual Installation Steps

### Add Helm Chart Dependencies
```bash
helm repo add dapr https://dapr.github.io/helm-charts/
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
helm repo add ratify https://deislabs.github.io/ratify
helm repo update
```

### Install Dapr
```bash
helm upgrade --install dapr dapr/dapr --namespace dapr-system --create-namespace --wait
```

### Install Gatekeeper
```bash
helm install gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=5 \
    --set mutatingWebhookTimeoutSeconds=2
```

### Install Redis
```bash
helm upgrade --install redis bitnami/redis --namespace gatekeeper-system --set image.tag="7.0-debian-11" --wait
```

Apply dapr state store encyrption secret using a generated key
```bash
SIGN_KEY=$(openssl rand 16 | hexdump -v -e '/1 "%02x"' | base64)

cat <<EOF > dapr-redis-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: ratify-dapr-signing-key
data:
  signingKey: $SIGN_KEY
EOF

kubectl apply -f dapr-redis-secret.yaml -n gatekeeper-system
```

Apply dapr state store Component custom resource
```bash
cat <<EOF > dapr-redis.yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: dapr-redis
spec:
  type: state.redis
  version: v1
  metadata:
  - name: redisHost
    value: redis-master:6379
  - name: redisPassword
    secretKeyRef:
      name: redis
      key: redis-password
  - name: primaryEncryptionKey
    secretKeyRef:
      name: ratify-dapr-signing-key
      key: signingKey
auth:
  secretStore: kubernetes
EOF

kubectl apply -f dapr-redis.yaml -n gatekeeper-system
```

### Install Ratify
```bash
# download the notation verification certificate
curl -sSLO https://raw.githubusercontent.com/deislabs/ratify/main/test/testdata/notation.crt
helm install ratify \
    ratify/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notationCert=./notation.crt \
    --set featureFlags.RATIFY_CERT_ROTATION=true \
    --set featureFlags.RATIFY_DAPR_CACHE_PROVIDER=true \
    --set replicaCount=3 \
	--set provider.cache.type="dapr" \
	--set provider.cache.name="dapr-redis"
```

### See Ratify in action

- Deploy a `demo` constraint.
```
kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
```

Once the installation is completed, you can test the deployment of an image that is signed using notation.

- This will successfully create the pod `demo`

```bash
kubectl run demo --image=ghcr.io/deislabs/ratify/notary-image:signed
kubectl get pods demo
```

Optionally you can see the output of the pod logs via: `kubectl logs demo`

- Now deploy an unsigned image

```bash
kubectl run demo1 --image=ghcr.io/deislabs/ratify/notary-image:unsigned
```

You will see a deny message from Gatekeeper denying the request to create it as the image doesn't have any signatures.

```bash
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: wabbitnetworks.azurecr.io/test/net-monitor:unsigned
```

You just validated the container images in your k8s cluster!

### Uninstall Ratify
Notes: Helm does NOT support upgrading CRDs, so uninstalling Ratify will require you to delete the CRDs manually. Otherwise, you might fail to install CRDs of newer versions when installing Ratify.
```bash
kubectl delete -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl delete -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
helm delete ratify --namespace gatekeeper-system
kubectl delete crd stores.config.ratify.deislabs.io verifiers.config.ratify.deislabs.io certificatestores.config.ratify.deislabs.io policies.config.ratify.deislabs.io
helm delete redis --namespace gatekeeper-system
helm delete dapr --namespace dapr-system
kubectl delete Component dapr-redis -n gatekeeper-system
kubectl delete Secret ratify-dapr-signing-key -n gatekeeper-system
helm delete gatekeeper -n gatekeeper-system
```

## Development Testing with Helmfile (build your own images)
While developing for HA scenarios, the `dev.high-availability.helmfile.yaml` can be useful.

Prerequisites:
- Install helmfile
- Build your own images (follow instructions [here](../../CONTRIBUTING.md#build-an-image-with-your-local-changes))
- Install Ratify + Gatekeeper on cluster with `dev.helmfile.yaml` (follow instructions [here](../../CONTRIBUTING.md#deploy-using-dev-helmfile))

### Update `dev.high-availability.helmfile.yaml`
Replace `repository`, `crdRepository`, and `tag` with previously built images:
```yaml
- name: image.repository 
  value: <YOUR RATIFY IMAGE REPOSITORY NAME>
- name: image.crdRepository
  value: <YOUR RATIFY CRD IMAGE REPOSITORY NAME>
- name: image.tag
  value: <YOUR IMAGES TAG NAME>
```

Deploy to cluster:
```bash
helmfile sync -f dev.high-availability.helmfile.yaml
```