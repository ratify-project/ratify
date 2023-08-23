# Automating Ratify on AKS and Azure Workload Identity with Terraform

This guide will walk you through the process of setting up and deploying the necessary Azure resources for Ratify using Terraform.

The deployment process involves the creation of various Azure resources, including a resource group, an Azure Key Vault instance, an Azure Container Registry, an Azure Kubernetes Service Cluster, and the configuration of Workload Identity for Kubernetes to authenticate with Azure resources. 

Once these resources are set up, you can deploy Ratify to Azure using Helm or other alternative installation methods.

```bash
                      +-------------------------+
                      |   Resource Group        |
                      +-------------------------+
                                |
            +----------------------------------------+
            |               |                        |
  +-----------------+   +-----------------+    +-------------------+
  | Azure Key Vault |   | Azure Container |    | Azure Kubernetes |
  |                 |   |   Registry      |    |      Service      |
  +-----------------+   +-----------------+    +-------------------+
                                                        |
                        +----------------------------------+
                        |                                  |
            +-------------------------+    +-----------------------------+
            | Azure User Assigned     |    |     Azure Federated         |
            | Managed Identity        |    |         Credential          |
            +-------------------------+    +-----------------------------+
```

## Prerequisites

| | |
|----------------------|------------------------------------------------------|
| GitHub account       | [Get a free GitHub account](https://github.com/join) |
| Azure account        | [Get a free Azure account](https://azure.microsoft.com/free) |
| Azure CLI            | [Install Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) |
| Terraform            | [Install Terraform](https://developer.hashicorp.com/terraform/tutorials/azure-get-started/install-cli) |

<div class="info" data-title="Note">

> The `terraform.tfvars` file that's provided contains example resource group names. Before continuing, please update the values with your desired Azure resource names.

</div>

## Log into Azure with the Azure CLI.

First, log into Azure with the Azure CLI.

```bash
az login
```

## Create a Service Principal

Next, create a Service Principal for Terraform to use to authenticate to Azure.

```bash
subscription_id=$(az account show --query id -o tsv)

az ad sp create-for-rbac --role="Contributor" --scopes="/subscriptions/$subscription_id"
```

<details>
<summary>Example Output</summary>

```output
{
  "appId": "00000000-0000-0000-0000-000000000000",
  "displayName": "azure-cli-2024-04-19-19-38-16",
  "name": "http://azure-cli-2024-04-19-19-38-16",
  "password": "QVexZdqvcPxx%4HJ^ZY",
  "tenant": "00000000-0000-0000-0000-000000000000"
}
```

</details>

Take note of the `appId`, `password`, and `tenant` values and store them in a secure location. You'll need them later.

## Assign the User Access Administrator role to the Service Principal

Next, assign the User Access Administrator role to the Service Principal. This role will allow Terraform to create the federated identity credential used by the workload identity.

```bash
az role assignment create --role "User Access Administrator" --assignee  "00000000-0000-0000-0000-000000000000" --scope "/subscriptions/$subscription_id"
```

Replace `00000000-0000-0000-0000-000000000000` with the `appId` value from the previous step.

## Export the Service Principal credentials as environment variables

Next, export the Service Principal credentials as environment variables. These variables will be used by Terraform to authenticate to Azure.

```bash
export ARM_CLIENT_ID="00000000-0000-0000-0000-000000000000"
export ARM_CLIENT_SECRET="QVexZdqvcPxx%4HJ^ZY"
export ARM_SUBSCRIPTION_ID=$subscription_id
export ARM_TENANT_ID="00000000-0000-0000-0000-000000000000"
```

Replace `00000000-0000-0000-0000-000000000000` with the `appId`, `password`, and `tenant` values from the previous step.

## Sign into Azure CLI with the Service Principal

Change the Azure CLI login from your user to the service principal you just created. This allows Terraform to consistently configure access polices to Azure Key Vault for the current user.

```bash
az login --service-principal -u $ARM_CLIENT_ID -p $ARM_CLIENT_SECRET --tenant $ARM_TENANT_ID
```

## Deploy the Terraform configuration

Next, deploy the Terraform configuration. This will create the Azure resources needed for this workshop.

```bash
cd terraform;
terraform init;
terraform apply
```

<details>
<summary>Example Output</summary>

```output
azurerm_resource_group.rg: Creating...
azurerm_resource_group.rg: Creation complete after 1s [id=/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg]
azurerm_key_vault.kv: Creating...
azurerm_key_vault.kv: Creation complete after 4s [id=https://kv.vault.azure.net]
azurerm_user_assigned_identity.ua: Creating...
azurerm_user_assigned_identity.ua: Creation complete after 1s [id=/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ua]
azurerm_container_registry.acr: Creating...
```

</details>


<div class="info" data-title="warning">

> Certain Azure resources need to be globally unique. If you receive an error that a resource already exists, you may need to change the name of the resource in the `terraform.tfvars` file.

</div>

## Export Terraform output as environment variables

As part of the Terraform deployment, several output variables were created. These variables will be used to configure the Ratify environment. 

Run the following command to export the Terraform output as environment variables:

<details>
<summary>Bash</summary>

```bash
export VAULT_URI="$(terraform output -raw key_vault_uri)"
export CERT_NAME="$(terraform output -raw ratify_certificate_name)"
export TENANT_ID="$(terraform output -raw tenant_id)"
export CLIENT_ID="$(terraform output -raw workload_identity_client_id)"
```

</details>

<details>

<summary>PowerShell</summary>

```pwsh
$VAULT_URI="$(terraform output -raw key_vault_uri)"
$CERT_NAME="$(terraform output -raw ratify_certificate_name)"
$TENANT_ID="$(terraform output -raw tenant_id)"
$CLIENT_ID="$(terraform output -raw workload_identity_client_id)"
```

</details>
