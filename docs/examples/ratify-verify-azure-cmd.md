# Enforce only signed images are allowed to be deployed on AKS with Notation and Ratify

The Azure Key Vault (AKV) is used to store a signing key that can be utilized by Notation with the Notation AKV plugin (azure-kv) to sign and verify container images and other artifacts. The Azure Container Registry (ACR) allows you to store and distribute signed images with signatures.

The signed containers enable users to assure deployments are built from a trusted entity and verify artifact hasn't been tampered with since their creation. The signed artifact ensures integrity and authenticity before the user pulls an artifact into any environment and avoid attacks. 

This article walks you through an end-to-end workflow of validating and enforcing only signed images are allowed to be deployed on AKS with Notation and Ratify.

In this article:


* Create and store a signing certificate in Azure Key Vault
* Sign a container image with notation
* Create an AKS cluster and ACR registry with Azure Workload Identity configured
* Intall Ratify and Gatekeeper
* Validate a container image signature with Ratify and Gatekeeper
* Deploy a signed image to AKS

![](https://i.imgur.com/1f9dfV9.png)


## Install the Notation CLI and AKV plugin

1. Install Notation v1.0.0-rc.2 on a Linux environment. You can also download the package for other environments from the [release page](https://github.com/notaryproject/notation/releases/tag/v1.0.0-rc.2).

    ```bash
    # Download, extract and install
    curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v1.0.0-rc.2/notation_1.0.0-rc.2_linux_amd64.tar.gz
    tar xvzf notation.tar.gz
            
    # Copy the notation cli to the desired bin directory in your PATH
    sudo cp ./notation /usr/local/bin
    ```

2. Install the notation Azure Key Vault plugin for remote signing and verification.

    > NOTE:
    > The plugin directory varies depending upon the operating system being used.  The directory path below assumes Ubuntu.
    > Please read the [notation config article](https://github.com/notaryproject/notation/blob/main/specs/notation-config.md) for more information.
    
    ```bash
    # Create a directory for the plugin
    mkdir -p ~/.config/notation/plugins/azure-kv
    
    # Download the plugin
    curl -Lo notation-azure-kv.tar.gz \
        https://github.com/Azure/notation-azure-kv/releases/download/v0.5.0-rc.1/notation-azure-kv_0.5.0-rc.1_Linux_amd64.tar.gz
    
    # Extract to the plugin directory
    tar xvzf notation-azure-kv.tar.gz -C ~/.config/notation/plugins/azure-kv notation-azure-kv
    ```

3. List the available plugins and verify that the plugin is available.

    ```bash
    notation plugin ls
    ```

## Configure environment variables

> [!NOTE]
> For easy execution of commands in the tutorial, provide values for the Azure resources to match the existing ACR and AKV resources.

1. Configure AKV resource names.

    ```bash
    # Name of the existing Azure Key Vault used to store the signing keys
    AKV_NAME=<your-unique-keyvault-name>
    # New desired key name used to sign and verify
    KEY_NAME=wabbit-networks-io
    CERT_PATH=./${KEY_NAME}.pem
    ```

2. Configure ACR and image resource names.

    ```bash
    # Name of the existing registry example: myregistry.azurecr.io
    ACR_NAME=<your-registry-name>
    # Existing full domain of the ACR
    REGISTRY=$ACR_NAME.azurecr.io
    # Container name inside ACR where image will be stored
    REPO=net-monitor
    TAG=v1
    IMAGE=$REGISTRY/${REPO}:$TAG
    # Source code directory containing Dockerfile to build
    IMAGE_SOURCE=https://github.com/wabbit-networks/net-monitor.git#main
    ```

## Store the signing certificate in AKV

If you have an existing certificate, upload it to AKV. For more information on how to use your own signing key, see the [signing certificate requirements.](https://github.com/notaryproject/notaryproject/blob/main/signature-specification.md#certificate-requirements)
Otherwise create an x509 self-signed certificate storing it in AKV for remote signing using the steps below.

### Create a self-signed certificate (Azure CLI)

1. Create a certificate policy file.

    Once the certificate policy file is executed as below, it creates a valid signing certificate compatible with **notation** in AKV. The EKU listed is for code-signing, but isn't required for notation to sign artifacts.

    ```bash
    cat <<EOF > ./my_policy.json
    {
        "issuerParameters": {
        "certificateTransparency": null,
        "name": "Self"
        },
        "x509CertificateProperties": {
        "ekus": [
        "1.3.6.1.5.5.7.3.3"
        ],
        "keyUsage": [
          "digitalSignature"
        ],
        "subject": "CN=wabbit-networks.io,O=Notary,L=Seattle,ST=WA,C=US",
        "validityInMonths": 12
        }
    }
    EOF
    ```

1. Create the certificate.

    ```azure-cli
    az keyvault certificate create -n $KEY_NAME --vault-name $AKV_NAME -p @my_policy.json
    ```

1. Get the Key ID for the certificate.

    ```bash
    KEY_ID=$(az keyvault certificate show -n $KEY_NAME --vault-name $AKV_NAME --query 'kid' -o tsv)
    ```
4. Download public certificate.

    ```bash
    CERT_ID=$(az keyvault certificate show -n $KEY_NAME --vault-name $AKV_NAME --query 'id' -o tsv)
    az keyvault certificate download --file $CERT_PATH --id $CERT_ID --encoding PEM
    ```

## Build and sign a container image

1. Build and push a new image with ACR Tasks.

    ```azure-cli
    az acr build -r $ACR_NAME -t $IMAGE $IMAGE_SOURCE
    ```

2. Authenticate with your individual Azure AD identity to use an ACR token.

    ```azure-cli
    export USER_NAME="00000000-0000-0000-0000-000000000000"
    export PASSWORD=$(az acr login --name $ACR_NAME --expose-token --output tsv --query accessToken)
    notation login -u $USER_NAME -p $PASSWORD $REGISTRY
    ```
3. Add a signing key referecing the Key ID

    ```bash
    notation key add $KEY_NAME --plugin azure-kv --id $KEY_ID
    ```

4. List the keys to confirm.

    ```bash
    notation key ls
    ```

5. Choose [COSE](https://datatracker.ietf.org/doc/html/rfc8152) signature format to sign the container image.

   - Sign the container image with the COSE signature envelope:
 
    ```bash
    notation sign --signature-format cose --key $KEY_NAME $IMAGE
    ```

## View the signed images associated with signatures

Signed images can be viewed with the `notation list` command

```bash
notation list $IMAGE
```
## Create and configure Azure Workload Identity

Ratify pulls artifacts from a private Azure Container Registry using Workload Federated Identity in an Azure Kubernetes Service cluster. For an overview on how workload identity operates in Azure, refer to the [documentation](https://docs.microsoft.com/en-us/azure/active-directory/develop/workload-identity-federation). You can use workload identity federation to configure an Azure AD app registration or user-assgined managed identity. The following workflow include Workload Identity configuration.

### Set up Workload Identity with AKS, ACR and AKV:
The official steps for setting up Workload Identity on AKS can be found here.

> Note: if you have identity authenticaton issue in your local machie, you can switch to use Azure Cloud Shell the complete this section.

1. Configure environment variables.

```bash
export IDENTITY_NAME=<Identity Name>
export GROUP_NAME=<Azure Resource Group Name>
export SUBSCRIPTION_ID=<Azure Subscription ID>
export TENANT_ID=<Azure Tenant ID>
export ACR_NAME=<Azure Container Registry Name>
export AKS_NAME=<Azure Kubernetes Service Name>
export KEYVAULT_NAME=<Azure Key Vault Name>
export RATIFY_NAMESPACE=<Namespace where Ratify deployed, defaults to "gatekeeper-system">
export CERT_DIR=<Directory storing TLS certs>
export NOTARY_PERM_NAME=<Name of cerficicate file uploaded to Key Vault>
```
2. Create a Workload Federated Identity.

```bash
az identity create --name "${IDENTITY_NAME}" --resource-group "${GROUP_NAME}" --location "${LOCATION}" --subscription "${SUBSCRIPTION_ID}"

export IDENTITY_OBJECT_ID="$(az identity show --name "${IDENTITY_NAME}" --resource-group "${GROUP_NAME}" --query 'principalId' -otsv)"
export IDENTITY_CLIENT_ID=$(az identity show --name ${IDENTITY_NAME} --resource-group ${GROUP_NAME} --query 'clientId' -o tsv)
```

3. Create an ACR registry and enable `AcrPull` role to the workload identity.

```bash
az acr create --name "${ACR_NAME}" \
  --resource-group "${GROUP_NAME}" \
  --sku Standard
```

4. Configure user-assigend managed identity

```bash
az role assignment create \
  --assignee-object-id ${IDENTITY_OBJECT_ID} \
  --role acrpull \
  --scope subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${GROUP_NAME}/providers/Microsoft.ContainerRegistry/registries/${ACR_NAME}
```

5. Create an OIDC enabled AKS cluster by following the steps below:

```bash
# Install the aks-preview extension
az extension add --name aks-preview

# Register the 'EnableWorkloadIdentityPreview' feature flag
az feature register --namespace "Microsoft.ContainerService" --name "EnableWorkloadIdentityPreview"
az provider register --namespace Microsoft.ContainerService

az aks create \
    --resource-group "${GROUP_NAME}" \
    --name "${AKS_NAME}" \
    --node-vm-size Standard_DS3_v2 \
    --node-count 1 \
    --generate-ssh-keys \
    --enable-workload-identity \
    --attach-acr ${ACR_NAME} \
    --enable-oidc-issuer

# Connect to the AKS cluster:
az aks get-credentials --resource-group ${GROUP_NAME} --name ${AKS_NAME}

export AKS_OIDC_ISSUER="$(az aks show -n ${AKS_NAME} -g ${GROUP_NAME} --query "oidcIssuerProfile.issuerUrl" -otsv)"
```

6. Establish federated identity credential. On AZ CLI `${RATIFY_NAMESPACE}` is where you deploy Ratify:

```bash
az identity federated-credential create \
  --name ratify-federated-credential \
  --identity-name "${IDENTITY_NAME}" \
  --resource-group "${GROUP_NAME}" \
  --issuer "${AKS_OIDC_ISSUER}" \
  --subject system:serviceaccount:"${RATIFY_NAMESPACE}":"gatekeeper-system"
```

7. Create an Azure Key Vault and set an access policy. If you have an AKV, you can skip this step.

```bash
az keyvault create \
  --resource-group ${GROUP_NAME} \
  --name ${KEYVAULT_NAME}
```

8. Set the environmental variable for Azure Key Vault URI.

```bash
export VAULT_URI=$(az keyvault show --name ${KEYVAULT_NAME} --resource-group ${GROUP_NAME} --query "properties.vaultUri" -otsv)
```

9. Import your own private key and certificates. You can import it on the portal as well.

```bash
az keyvault certificate import \
  --vault-name ${KEYVAULT_NAME} \
  -n <Certificate Name> \
  -f /path/to/certificate
```
 
10. Configure policy for user-assigned managed identity:
    
```bash
az keyvault set-policy --name ${KEYVAULT_NAME} \
  --certificate-permissions get \
  --object-id ${IDENTITY_OBJECT_ID}
```

## Deploy Gatekeeper and Ratify on AKS 

1. Deploy Gatekeeper from helm chart:

```bash
helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts

helm install gatekeeper/gatekeeper  \
  --name-template=gatekeeper \
  --namespace gatekeeper-system --create-namespace \
  --set enableExternalData=true \
  --set validatingWebhookTimeoutSeconds=5 \
  --set mutatingWebhookTimeoutSeconds=2
```

2. Deploy Ratify from helm chart:

```bash
helm clone https://github.com/deislabs/ratify.git

helm install ratify ./charts/ratify --atomic \
  --namespace ${RATIFY_NAMESPACE} \
  --set-file provider.tls.crt=${CERT_DIR}/server.crt \
  --set-file provider.tls.key=${CERT_DIR}/server.key \
  --set provider.tls.cabundle="$(cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
  --set oras.authProviders.azureWorkloadIdentityEnabled=true \
  --set azureWorkloadIdentity.clientId=${IDENTITY_CLIENT_ID} \
  --set akvCertConfig.enabled=true \
  --set akvCertConfig.vaultURI=${VAULT_URI} \
  --set akvCertConfig.cert1Name=${NOTARY_PEM_NAME} \
  --set akvCertConfig.tenantId=${TENANT_ID}
```

3. Enforce Gatekeeper policy to allow only signed images can be deployed:

```bash
kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
```

4. This step above may take around 10 minutes to complete and registration status can be checked by running the following command:

```bash
az feature show --namespace "Microsoft.ContainerService" --name "EnableWorkloadIdentityPreview" -o table
Name                                                      RegistrationState
--------------------------------------------------------    -------------------
Microsoft.ContainerService/EnableWorkloadIdentityPreview    Registered
```

## Deploy an signed image to AKS cluster

1. Deploy an signed image to AKS cluster. It can be deployed to the AKS cluster.

```bash
$ kubectl run ratify-demo-signed --image=$IMAGE
Pod ratify-demo-signed created
```

2. Deploy an unsigned image to AKS cluster. The deployment has been denied since the image doesn't meet the deployment criteria. 

```bash
$ kubectl run ratify-demo-unsigned --image=unsigned:v1
Error from server (Forbidden): admission webhook "validation.gatekeeper.sh" denied the request: [ratify-constraint] Subject failed verification: wabbitnetworks.azurecr.io/test/net-monitor:unsigned
```
