# Dynamic Plugins

While Ratify provides a number of built-in referrer stores and verifiers, there may be times that you need to add additional out-of-tree plugins to verify additional artifact types or check for specific conditions. Whether you've followed the [plugin authoring guide](../quickstarts/creating-plugins.md), or you want to use a plugin that's provided by a third party, the next step is to make your plugin available to be called at runtime.

Ratify includes an optional dynamic plugins feature to simplify plugin distribution. This removes the need to rebuild or maintain your own Ratify container image. When enabled, plugins are stored as OCI artifacts and pulled down at runtime when they are registered via CRD. This guide will show you how to activate and configure dynamic plugins.

## Uploading your plugin

Upload your plugin to a registry that supports OCI artifacts. Here we're using the sample plugin located at `plugins/verifier/sample`. Replace the registry values with your own.

```shell
CGO_ENABLED=0 go build -o=./sample ./plugins/verifier/sample
oras push myregistry.azurecr.io/sample-plugin:v1 ./sample
```

> It's important to build plugins with `CGO_ENABLED=0` so that they can run inside the Ratify [distroless](https://github.com/GoogleContainerTools/distroless) container

## Ratify Configuration

Now that you have a plugin available as an OCI artifact, it's time to enable the `RATIFY_EXPERIMENTAL_DYNAMIC_PLUGINS` [feature flag](../reference/usage.md#feature-flags). Here, we enable it via Helm chart parameter:

```shell
# Option 1: to enable on a fresh install
helm install ratify \
    ratify/ratify --atomic \
    --set-file provider.tls.crt=${CERT_DIR}/server.crt \
    --set-file provider.tls.key=${CERT_DIR}/server.key \
    --set provider.tls.cabundle="$(cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
    --set featureFlags.RATIFY_EXPERIMENTAL_DYNAMIC_PLUGINS=true

# Option 2: to enable on a previously-installed release
helm upgrade ratify \
  ratify/ratify --atomic \
  --reuse-values \
  --set featureFlags.RATIFY_EXPERIMENTAL_DYNAMIC_PLUGINS=true
```

## Plugin Configuration

The last step is to specify the optional `source` property to tell Ratify where to download the plugin from. Dynamic plugins use the same [auth providers](../reference/oras-auth-provider.md) and options as the builtin ORAS store (ex: `azureWorkloadIdentity`, `awsEcrBasic`, `k8Secrets`) for authentication to the registry where your plugins are located.

```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Verifier
metadata:
  name: verifier-sample
spec:
  name: sample
  artifactTypes: application/vnd.ratify.spdx.v0
  source:
    artifact: myregistry.azurecr.io/sample-plugin:v1
    authProvider:
      name: azureWorkloadIdentity
```

## Confirmation / Troubleshooting

You can check the Ratify logs for more details on which plugin(s) were downloaded. Your specific commands may vary slightly based on the values you provided during chart installation.

```shell
kubectl logs -n gatekeeper-system deployment/ratify
```

This will generate output similar to below, which can be used for confirmation of a successful plugin download or to aid in troubleshooting.

```text
time="2023-01-18T16:44:46Z" level=info msg="Setting log level to info"
time="2023-01-18T16:44:46Z" level=info msg="Feature flag EXPERIMENTAL_DYNAMIC_PLUGINS is enabled"
time="2023-01-18T16:44:46Z" level=info msg="starting crd manager"
time="2023-01-18T16:44:46Z" level=info msg="initializing executor with config file at default config path"
<snip>
time="2023-01-18T16:44:46Z" level=info msg="reconciling verifier 'verifier-dynamic'"
time="2023-01-18T16:44:46Z" level=info msg="Address was empty, setting to default path: /.ratify/plugins"
time="2023-01-18T16:44:47Z" level=info msg="selected auth provider: azureWorkloadIdentity"
time="2023-01-18T16:44:48Z" level=info msg="downloaded verifier plugin dynamic from myregistry.azurecr.io/sample-plugin:v1 to /.ratify/plugins/dynamic"
time="2023-01-18T16:44:48Z" level=info msg="verifier 'dynamic' added to verifier map"
<snip>
```
