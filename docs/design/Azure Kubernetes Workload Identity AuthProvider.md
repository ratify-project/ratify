# Azure Kubernetes Workload Identity AuthProvider
Author: Akash Singhal (@akashsinghal)

Pod Identity will be deprecated in the future and the industry standard moving forward for accessing cloud services from an application is [Workload Identity](https://azure.github.io/azure-workload-identity/docs/introduction.html).

User would need to follow instructions [here](https://azure.github.io/azure-workload-identity/docs/quick-start.html) to enable workload identity for their cluster. To access a specific ACR, the AAD Application created for Workload Identity would need to be assigned an ACRpull role to the ACR. 

## High Level Auth Flow using Workload Identity

1. The Kubernetes Service Account, which is provided federated identity, injects the tenant_id, authority_host, client_id, and token_file path as environment variables in the pod. The token file is also mounted at the token_file path.
2. Token file is used to generate a confidential client (msal-go library)
3. Confidential Client is used to request an AAD token with the `https://containerregistry.azure.com/` scope. This returns a valid containerregistry scoped access token.
4. Artifact login server is challenged and a directive is returned for token exchange.
5. ARM token is exchanged for a registry refresh token
6. Registry username is simply the default docker GUID (`00000000-0000-0000-0000-000000000000`) and the password is the registry refresh token


## User steps to setup an AKS cluster and delegate access to specific private ACR
The official steps for setting up Workload Identity on AKS can be found [here](https://azure.github.io/azure-workload-identity/docs/quick-start.html).  

1. Create ACR
2. Create OIDC enabled AKS cluster by follow steps [here](https://learn.microsoft.com/en-us/azure/aks/use-oidc-issuer#create-an-aks-cluster-with-oidc-issuer)
3. Save the cluster's OIDC URL: `az aks show --resource-group <resource_group> --name <cluster_name> --query "oidcIssuerProfile.issuerUrl" -otsv`
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
  containers:
    - image: ghcr.io/azure/azure-workload-identity/msal-go:latest
      name: oidc
      env:
      - name: KEYVAULT_NAME
        value: ${KEYVAULT_NAME}
      - name: SECRET_NAME
        value: ${KEYVAULT_SECRET_NAME}
  nodeSelector:
    kubernetes.io/os: linux
EOF
```

## AzureWIAuthProvider Implementation
```
// AzureK8Conf describes the configuration of Azure K8s Auth Provider
type AzureWIConf struct {
    Name           string     `json:"name"`
}

type AzureWIAuthProviderFactory struct{}

type AzureWIAuthProvider struct {}

// init calls Register for our Azure WI provider
func init() {
    AuthProviderFactory.Register("azure-wi", &AzureWIAuthProviderFactory)
}

// create returns the AzureK8AuthProvider
func (s *AzureWIAuthProviderFactory) Create(authProviderConfig AuthProviderConfig) (AuthProvider, error)

// Enabled implements AuthProvider; Can be used to verify all fields for WI exist
func (d *AzureK8AuthProvider) Enabled() bool

// Provide implements AuthProvider; follows auth flow outlined above
func (d *AzureWIAuthProvider) Provide(artifact string) (AuthConfig, error) 
```
## Questions
1. The Gatekeeper timeout of 3 seconds leads to a premature timeout of the deployment. The authentication requires many roundtrips. Currently, a basic registry client cache is implemented such that a retry of the deployment will succeed in the next attempt. Is there a way to increase Gatekeeper timeout? Or frontload Auth before the webhook starts?