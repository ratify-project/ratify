# ORAS Authentication Provider

ORAS handles all referrer operations using registry as the referrer store. It uses authentication credentials to authenticate to a registry and access referrer artifacts. Ratify contains many Authentication Providers to support different authentication scenarios. The user specifies which authentication provider to use in the configuration.

The `authProvider` section of configuration file specifies the authentication provider. The `name` field is REQUIRED for Ratify to bind to the correct provider implementation. 

## Example config.json
```
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
                "application/vnd.cncf.notary.v2.signature": "any"
            }
        }
    },
    "verifier": {
        "version": "1.0.0",
        "plugins": [
            {
                "name":"notaryv2",
                "artifactTypes" : "application/vnd.cncf.notary.v2.signature",
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

### 1. Docker Config
This is the default authentication provider. Ratify attempts to look for credentials at the default docker configuration path ($HOME/.docker/config.json) if the `authProvider` section is not specified.

Specify the `configPath` field for the `dockerConfig` authentication provider to use a different docker config file path. 

```
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

### 2. Azure Workload Identity
Ratify pulls artifacts from a private Azure Container Registry using Workload Federated Identity in an Azure Kubernetes Service cluster. For an overview on how workload identity operates in Azure, refer to the [documentation](https://docs.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation). 

#### User steps to set up Workload Identity with AKS and ACR:

The official steps for setting up Workload Identity on AKS can be found [here](https://azure.github.io/azure-workload-identity/docs/quick-start.html).  

1. Create ACR
2. Create OIDC enabled AKS cluster by following steps [here](https://docs.microsoft.com/en-us/azure/aks/cluster-configuration#oidc-issuer-preview)
3. Connect to cluster using kubectl by following steps [here](https://docs.microsoft.com/en-us/azure/aks/tutorial-kubernetes-deploy-cluster?tabs=azure-cli#connect-to-cluster-using-kubectl)
4. Save the cluster's SERVICE_ACCOUNT_ISSUER (OIDC url): `az aks show --resource-group <resource_group> --name <cluster_name> --query "oidcIssuerProfile.issuerUrl" -otsv`
5. Install Mutating Admission Webhook onto AKS cluster by following steps [here](https://azure.github.io/azure-workload-identity/docs/installation/mutating-admission-webhook.html)
6. As the guide linked above shows, it's possible to use the AZ workload identity CLI or the regular az CLI to perform remaining setup. Following steps follow the AZ CLI.
7. Create ACR AAD application: `az ad sp create-for-rbac --name "<APPLICATION_NAME>"`
8. On Portal or AZ CLI, enable acrpull role to the AAD application for the ACR resource
```
// Sample AZ CLI command
az role assignment create --assignee <AAD APPLICATION ID> --role acrpull --scope subscriptions/<SUBSCRIPTION NAME>/resourceGroups/<RESOURCE GROUP>/providers/Microsoft.ContainerRegistry/registries/<REGISTRY NAME>
```
9. Use kubectl to add service account to cluster: 
```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    azure.workload.identity/client-id: ${APPLICATION_CLIENT_ID}
  labels:
    azure.workload.identity/use: "true"
  name: ${SERVICE_ACCOUNT_NAME}
  namespace: ${SERVICE_ACCOUNT_NAMESPACE}
EOF
```
10. From Azure Cloud Shell: 
```
export APPLICATION_OBJECT_ID="$(az ad app show --id ${APPLICATION_CLIENT_ID} --query objectId -otsv)"
cat <<EOF > body.json
{
  "name": "kubernetes-federated-credential",
  "issuer": "${SERVICE_ACCOUNT_ISSUER}",
  "subject": "system:serviceaccount:${SERVICE_ACCOUNT_NAMESPACE}:${SERVICE_ACCOUNT_NAME}",
  "description": "Kubernetes service account federated credential",
  "audiences": [
    "api://AzureADTokenExchange"
  ]
}
EOF

az rest --method POST --uri "https://graph.microsoft.com/beta/applications/${APPLICATION_OBJECT_ID}/federatedIdentityCredentials" --body @body.json
```
11. In the pod template defintion, add the `serviceAccountName` property. Example Pod definition:
```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: quick-start
  namespace: <SERVICE_ACCOUNT_NAMESPACE>
spec:
  serviceAccountName: <SERVICE_ACCOUNT_NAME>
  ...
EOF
```

#### Ratify Auth Provider Configuration
```
"store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache",
                "authProvider": {
                    "name": "azureWorkloadIdentity"
                }
            }
        ]
    }
```

### 3. Kubernetes Secrets
Ratify resolves registry credentials from [Docker Config Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/#docker-config-secrets) in the cluster. Ratify considers kubernetes secrets in two ways:
1. The configuration can specify a list of `secrets`. Each entry REQUIRES the `secretName` field. The `namespace` field MUST also be provided if the secret does not exist in the namespace Ratify is deployed in. The Ratify helm chart contains a [roles.yaml](https://github.com/deislabs/ratify/blob/main/charts/ratify/templates/roles.yaml) file with role assignments. If a namespace other than Ratify's namespace is provided in the secret list, the user MUST add a new role binding to the cluster role for that new namespace.

2. Ratify considers the `imagePullSecrets` specified in the service account associated with Ratify. The `serviceAccountName` field specifies the service account associated with Ratify. Ratify MUST be assigned a role to read the service account and secrets in the Ratify namespace.

Ratify only supports the kubernetes.io/dockerconfigjson secret type or the legacy kubernetes.io/dockercfg type.  

#### Sample Configuration
```
"store": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache",
                "authProvider": {
                    "name": "k8Secrets",
                    "serviceAccountName": "ratify-sa", // will be 'default' if not specified
                    "secrets" : [
                        {
                            "secretName": "artifact-pull-dockerConfig" // Ratify namespace will be used 
                        },
                        {
                            "secretName": "artifact-pull-dockerConfig2",
                            "namespace": "test"
                        }
                    ]
                }
            }
        ]
    }
```

Note: Kubernetes secrets are reloaded and refreshed for Ratify to use every 12 hours. Changes to the Secret may not be reflected immediately. 

## Notational Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be interpreted as described in [RFC 2119](http://tools.ietf.org/html/rfc2119).

The key words "unspecified", "undefined", and "implementation-defined" are to be interpreted as described in the [rationale for the C99 standard](http://www.open-std.org/jtc1/sc22/wg14/www/C99RationaleV5.10.pdf#page=18).
