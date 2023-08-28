# ORAS Authentication Provider

ORAS handles all referrer operations using registry as the referrer store. It uses authentication credentials to authenticate to a registry and access referrer artifacts. Ratify contains many Authentication Providers to support different authentication scenarios. The user specifies which authentication provider to use in the configuration.

The `authProvider` section of configuration file specifies the authentication provider. The `name` field is REQUIRED for Ratify to bind to the correct provider implementation.

## Example config.json

```json
{
    "store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache",
                "authProvider": {
                    "name": "<auth provider name>",
                    <other provider specific fields>
                }
            }
        ]
    },
    "policy": {
        "version": "1.0.0",
        "plugin": {
            "name": "configPolicy",
            "artifactVerificationPolicies": {
                "application/vnd.cncf.notary.signature": "any"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [
            {
                "name":"notation",
                "artifactTypes" : "application/vnd.cncf.notary.signature",
                "verificationCerts": [
                    "<cert folder>"
                  ]
            }  
        ]
    }
}
```

NOTE: ORAS will attempt to use anonymous access if the authentication provider fails to resolve credentials.

## Supported Providers

1. Docker Config file
1. Azure Workload Identity
1. Azure Managed Identity
1. Kubernetes Secrets
1. AWS IAM Roles for Service Accounts (IRSA)

### 1. Docker Config

This is the default authentication provider. Ratify attempts to look for credentials at the default docker configuration path ($HOME/.docker/config.json) if the `authProvider` section is not specified.

Specify the `configPath` field for the `dockerConfig` authentication provider to use a different docker config file path.

```json
"store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache",
                "authProvider": {
                    "name": "dockerConfig",
                    "configPath": <custom file path string>
                }
            }
        ]
    }
```

Both the above modes uses a k8s secret of type ```dockerconfigjson``` that is described in the [document](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)

### 2. Azure Workload Identity

Ratify pulls artifacts from a private Azure Container Registry using Workload Federated Identity in an Azure Kubernetes Service cluster. For an overview on how workload identity operates in Azure, refer to the [documentation](https://docs.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation). You can use workload identity federation to configure an Azure AD app registration or user-assgined managed identity. 

Please refer to [quick start](https://github.com/deislabs/ratify/blob/main/docs/quickstarts/ratify-on-azure.md#configure-workload-identity-for-acr) to configure workload identity for ACR.
The official steps for setting up Workload Identity on AKS can be found [here](https://azure.github.io/azure-workload-identity/docs/quick-start.html).  

### 3. Kubernetes Secrets

Ratify resolves registry credentials from [Docker Config Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/#docker-config-secrets) in the cluster. Ratify considers kubernetes secrets in two ways:

1. The configuration can specify a list of `secrets`. Each entry REQUIRES the `secretName` field. The `namespace` field MUST also be provided if the secret does not exist in the namespace Ratify is deployed in. The Ratify helm chart contains a [ratify-manager-role-clusterrole.yaml](https://github.com/deislabs/ratify/blob/main/charts/ratify/templates/ratify-manager-role-clusterrole.yaml) file with role assignments. If a namespace other than Ratify's namespace is provided in the secret list, the user MUST add a new role binding to the cluster role for that new namespace.

2. Ratify considers the `imagePullSecrets` specified in the service account associated with Ratify. The `serviceAccountName` field specifies the service account associated with Ratify. Ratify MUST be assigned a role to read the service account and secrets in the Ratify namespace.

Ratify only supports the kubernetes.io/dockerconfigjson secret type.

> Note: Kubernetes secrets are reloaded and refreshed for Ratify to use every 12 hours. Changes to the Secret may not be reflected immediately.

#### Sample install guide using service account image pull secrets
- Create a k8s secret by providing credentials on the command line. This secret should be in the same namespace that contains Ratify deployment.

```bash
kubectl create secret docker-registry ratify-regcred -n <ratify-namespace> --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

- Deploy Ratify using helm

```bash
helm install ratify \
    ratify/ratify --atomic \
    --namespace gatekeeper-system \
    --set-file notationCert=./notation.crt \
    --set oras.authProviders.k8secretsEnabled=true
```

- Patch the Ratify service account (default is `ratify-admin`) for the ratify namespace (default is `gatekeeper-system`) to append `ratify-regcred` secret to `imagePullSecrets` array

```bash
kubectl patch serviceaccount ratify-admin -n gatekeeper-system -p '{"imagePullSecrets": [{"name": "ratify-regcred"}]}'
```

> This mode can be used to authenticate with a single registry. If authentication to multiple registries is needed, docker config file can be used as described below

If Docker config file is used for the registry login process, the same file can be used to create a k8s secret.

- Deploy Ratify by specifying the path to the Docker config file.

> Note: If you use a Docker credentials store, you won't see that auth entry but a credsStore entry with the name of the store as value. In such cases, this option cannot be used. 

```bash
helm install ratify charts/ratify --set-file dockerConfig=<path to the local Docker config file>
```

Both the above modes uses a k8s secret of type ```dockerconfigjson``` that is described in the [document](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)

#### Sample Store Resource with specified k8s secrets

```yaml
apiVersion: config.ratify.deislabs.io/v1beta1
kind: Store
metadata:
  name: store-oras
spec:
  name: oras
  parameters: 
    cacheEnabled: true
    ttl: 10
    useHttp: true  
    authProvider:
      name: k8Secrets
      serviceAccountName: ratify-admin
      secrets: 
      - secretName: artifact-pull-dockerConfig
      - secretName: artifact-pull-dockerConfig2
        namespace: test 
```

### 5. AWS IAM Roles for Service Accounts (IRSA)

Ratify pulls artifacts from a private Amazon Elastic Container Registry (ECR) using an ECR auth token. This token is accessed using the federated workload identity assigned to pods via [IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html). The AWS IAM Roles for Service Accounts Basic Auth provider uses the [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) to retrieve basic auth credentials based on a role assigned to a Kubernetes Service Account referenced by a pod specification. For a specific example of how IAM Roles for Service Accounts, a.k.a. IRSA, works with pods running the AWS SDK for Go v2, please see this [post](https://blog.jimmyray.io/kubernetes-workload-identity-with-aws-sdk-for-go-v2-927d2f258057).

#### User steps to set up IAM Roles for Service Accounts with Amazon EKS to access Amazon ECR

The official steps for setting up IAM Roles for Service Accounts with Amazon EKS can be found [here](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts-technical-overview.html).

1. Create an OIDC enabled Amazon EKS cluster by following steps [here](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html).
2. Connect to cluster using `kubectl` by following steps [here](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html)
3. Create an AWS Identity and Access Management (IAM) policy with permissions needed for Ratify to access Amazon ECR. Please see the official [documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_create.html) for creating AWS IAM policies. The AWS managed policy `arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly` will work for this purpose.
4. Create the `ratify` Namespace in your Amazon EKS cluster.
5. Using [eksctl](https://eksctl.io/usage/iamserviceaccounts/), create a Kubernetes Service Account that uses the policy from above.

```shell
eksctl create iamserviceaccount \
    --name ratify-admin \
    --namespace ratify \
    --cluster ratify \
    --attach-policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly \
    --approve \
    --override-existing-serviceaccounts
```

6. Verify that the Service Account was successfully created and annotated with a newly created role.

```shell
kubectl -n ratify get sa ratify -oyaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: <IAM_ROLE_ARN>
  labels:
    app.kubernetes.io/managed-by: eksctl
  name: ratify-admin
  namespace: ratify
secrets:
- name: ratify-token-...
```

___Note__: The creation of the role was done in the background by `eksctl` and an AWS CloudFormation stack, created by `eksctl`._ 

7. The Service Account created above should be referenced in the helm chart values, without creating the Service Account.

```yaml
serviceAccount:
  create: false
  name: ratify-admin
```

8. Specify the _AWS ECR Basic Auth_ provider in the Ratify helm chart [values](https://github.com/deislabs/ratify/blob/main/charts/ratify/values.yaml) file.

```yaml
oras:
  authProviders:
    azureWorkloadIdentityEnabled: false
    azureManagedIdentityEnabled: false
    k8secretsEnabled: false
    awsEcrBasicEnabled: true
    awsApiOverride:
      enabled: false
      endpoint: ""
      partition: "" # defaults to aws
      region: ""
```

9. If your AWS environment requires you to use a custom AWS endpoint resolver then you need to enable this feature in the helm chart [values](https://github.com/deislabs/ratify/blob/main/charts/ratify/values.yaml) file.

```yaml
oras:
  authProviders:
    azureWorkloadIdentityEnabled: false
    azureManagedIdentityEnabled: false
    k8secretsEnabled: false
    awsEcrBasicEnabled: true
    awsApiOverride:
      enabled: true
      endpoint: <AWS_ENDPOINT>
      partition: <AWS_PARTITION> # defaults to aws
      region: <AWS_REGION>
```

Once ratify is started, if an AWS custom endpoint resolver is successfully enabled, you will see the following log entries in the ratify pod logs, with no following errors:

```
AWS ECR basic auth using custom endpoint resolver...
AWS ECR basic auth API override endpoint: <AWS_ENDPOINT>
AWS ECR basic auth API override partition: <AWS_PARTITION>
AWS ECR basic auth API override region: <AWS_REGION>
```

10. [Install Ratify](https://github.com/deislabs/ratify#quick-start)

```shell
helm install ratify \
    ratify/ratify --atomic \
    --namespace ratify --values values.yaml
```

11. After install, verify that the Service Account is referenced by the `ratify` pod(s).

```shell
kubectl -n ratify get pod ratify-... -oyaml | grep serviceAccount
  serviceAccount: ratify-admin
  serviceAccountName: ratify-admin
      - serviceAccountToken:
      - serviceAccountToken:
```

12. Verify that the [Amazon EKS Pod Identity Webhook](https://github.com/aws/amazon-eks-pod-identity-webhook) created the environment variables, projected volumes, and volume mounts for the Ratify pod(s). 

```shell
kubectl -n ratify get po ratify-... -oyaml
...
    - name: AWS_STS_REGIONAL_ENDPOINTS
      value: regional
    - name: AWS_DEFAULT_REGION
      value: us-east-2
    - name: AWS_REGION
      value: us-east-2
    - name: AWS_ROLE_ARN
      value: <AWS_ROLE_ARN>
    - name: AWS_WEB_IDENTITY_TOKEN_FILE
      value: /var/run/secrets/eks.amazonaws.com/serviceaccount/token
...
    volumeMounts:
...
    - mountPath: /var/run/secrets/eks.amazonaws.com/serviceaccount
      name: aws-iam-token
      readOnly: true
...
  volumes:
  - name: aws-iam-token
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          audience: sts.amazonaws.com
          expirationSeconds: 86400
          path: token
...
```

13. Verify the _AWS ECR Basic Auth_ provider is configured in the `ratify-configuration` ConfigMap.

```shell
kubectl -n ratify get cm ratify-configuration -oyaml
...
"stores": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache",
                "auth-provider": {
                    "name": "awsEcrBasic"
                }
            }
        ]
    }
...
```

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).
