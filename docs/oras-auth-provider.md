# ORAS Authentication Provider

The ORAS referrer store is responsible for all registry and referrer resolution operations. In order for the ORAS referrer store to access referrer artifacts, authentication credentials must be provided to ORAS. Ratify contains multiple Authentication Providers to support different authentication scenarios within or outside of Kubernetes clusters. The user simply specifies which authentication provider to use in the configuration.

In the configuration file, the user specifies the authentication provider in the `auth-provider` section. The `name` field is **required** to bind ratify to the correct provider implementation. 

## Example config.json
```
{
    "stores": {
        "version": "1.0.0",
        "plugins": [
            {
                "name": "oras",
                "localCachePath": "./local_oras_cache",
                "auth-provider": {
                    "name": "<auth provider name>",
                    <other provider specific fields>
                }
            }
        ]
    },
    "verifiers": {
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


NOTE: If the authentication provider cannot resolve credentials for the specified registry host name, it will attempt to use anonymous credentials.

## Supported Providers

### 1. Docker Config
This is the default authentication provider. If the `auth-provider` section is not specified in the configuration file, ratify will attempt to look for credentials at the default docker configuration path ($HOME/.docker/config.json).

To specify a different docker config file path, the `docker-config` authentication provider can be defined in the configuration with the `configPath` field set. 

```
"auth-provider": {
    "name": "docker-config",
    "configPath": <custom file path string>
}
```

### 2. Azure Workload Identity
Using Workload Federated Identity in an Azure Kubernetes Service cluster, Ratify can pull artifacts from a private Azure Container Registry for verification. For an overview on how workload identity operates in Azure, refer to the [documentation](https://docs.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation). 

#### User steps to set up Workload Identity with AKS and ACR:

The official steps for setting up Workload Identity on AKS can be found [here](https://azure.github.io/azure-workload-identity/docs/quick-start.html).  

1. Create ACR
2. Create OIDC enabled AKS cluster by following steps [here](https://docs.microsoft.com/en-us/azure/aks/cluster-configuration#oidc-issuer-preview)
3. Save the cluster's SERVICE_ACCOUNT_ISSUER (OIDC url): `az aks show --resource-group <resource_group> --name <cluster_name> --query "oidcIssuerProfile.issuerUrl" -otsv`
4. Install Mutating Admission Webhook onto AKS cluster by following steps [here](https://azure.github.io/azure-workload-identity/docs/installation/mutating-admission-webhook.html)
5. As the guide linked above shows, it's possible to use the AZ workload identity CLI or the regular az CLI to perform remaining setup. Following steps follow the AZ CLI.
6. Create ACR AAD application: `az ad sp create-for-rbac --name "<APPLICATION_NAME>"`
7. On Portal or AZ CLI, enable acrpull role to the AAD application for the ACR resource
8. Use kubectl to add service account to cluster: 
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
9. From Azure Cloud Shell: 
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
10. In the Pod spec, add the `serviceAccountName` property. Example Pod spec:
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
"auth-provider": {
    "name": "azure-wi"
}
```

### 3. Kubernetes Secrets
Registry credentials can be extracted from kubernetes secrets. Users can specify Docker Config Kubernetes secrets that exist in the cluster, and Ratify will resolve registry credentials using these secrets. The kuberentes secrets Ratify will consider are specified in two ways. First, a list of `secrets` can be specified in the configuration. Each entry will require the `secretName` to be specifed. If the secret does not exist in the namespace ratify is deployed in, the `namespace` field must also be provided. Second, the `imagePullSecrets` specified in the service account associated with Ratify will also be considered. The service account associated with Ratify can be specifed using the `serviceAccountName` field.

Ratify only supports the kubernetes.io/dockerconfigjson secret type or the legacy kubernetes.io/dockercfg type.  

#### Sample Configuration
```
"auth-provider": {
    "name": "k8s-secrets",
    "serviceAccountName": "ratify-sa", // will be 'default' if not specified
    "secrets" : [
        {
            "secretName": "artifact-pull-docker-config" // ratify namespace will be used 
        },
        {
            "secretName": "artifact-pull-docker-config2",
            "namespace": "test"
        }
    ]
}
```

NOTE: The provided Ratify Helm Chart consists of a roles.yaml file which establishes the necessary cluster role and role binding to the ratify namespace in order for ratify to read secrets and get the service account. If a namespace other than Ratify's namespace is provided, the user must add a new role binding to the cluster role so ratify's service account can operate. 
