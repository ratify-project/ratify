#!/bin/bash

echo "Starting Ratify on Azure\n"
echo "RESOURCE_GROUP: $RESOURCE_GROUP\n"
echo "CLUSTER_NAME: $CLUSTER_NAME\n"
echo "ENABLE_MUTATION: $ENABLE_MUTATION\n"
echo "ENABLE_CERT_ROTATION: $ENABLE_CERT_ROTATION\n"
echo "MANAGED_IDENTITY_CLIENT_ID: $MANAGED_IDENTITY_CLIENT_ID\n"
echo "TRUSTED_IDENTITIES: $TRUSTED_IDENTITIES\n"
echo "REGISTRY_SCOPE: $REGISTRY_SCOPE\n"
echo "ROOT_CERT_URL: $ROOT_CERT_URL\n"

# File names for the certificates
CA_CRT_FILE="ca.crt"
CA_PEM_FILE="ca.pem"
TSA_CRT_FILE="tsa.crt"
TSA_PEM_FILE="tsa.pem"

echo "Downloading CA certificate from $ROOT_CERT_URL..."
curl -s -o $CA_CRT_FILE $ROOT_CERT_URL
if [ -f "$CA_CRT_FILE" ]; then
    echo "$CA_CRT_FILE was downloaded successfully."
else
    echo "$CA_CRT_FILE download failed."
    exit 1
fi

echo "Converting the CA certificate..."
openssl x509 -in "$CA_CRT_FILE" -inform DER -out "$CA_PEM_FILE" -outform PEM
if [ -f "$CA_PEM_FILE" ]; then
    echo "$CA_PEM_FILE was created successfully."
else
    echo "$CA_PEM_FILE creation failed."
    exit 1
fi

# Download and process the TSA certificate (optional)
if [[ -n "$CERT_URL_TSA" ]]; then
    echo "Downloading TSA certificate from $CERT_URL_TSA..."
    curl -s -o $TSA_CRT_FILE $CERT_URL_TSA
    openssl x509 -in $TSA_CRT_FILE -out $TSA_PEM_FILE -outform PEM
    TSA_HELM_ARG="--set-file notationCerts[1]=$TSA_PEM_FILE --set notation.trustPolicies[0].trustStores[1]=ca:notationCerts[1]"
else
    echo "TSA certificate URL not provided. Skipping TSA certificate configuration."
    TSA_HELM_ARG=""
fi

echo "TSA_HELM_ARG: $TSA_HELM_ARG"

# Get AKS credentials
az aks get-credentials --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME --overwrite-existing

# Get the client id
clientId=$(az aks show --resource-group $RESOURCE_GROUP --name $CLUSTER_NAME --query "identityProfile.kubeletidentity.clientId" -o tsv)
echo "clientId: $clientId"

# Install Helm
curl -s https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash 
# Add a Helm repo
helm repo add ratify https://ratify-project.github.io/ratify

# Install Ratify
helm install ratify ratify/ratify --atomic \
  --namespace gatekeeper-system --create-namespace \
  --set azureWorkloadIdentity.clientId=$clientId \
  --set provider.enableMutation=$ENABLE_MUTATION \
  --set featureFlags.RATIFY_CERT_ROTATION=$ENABLE_CERT_ROTATION \
  --set-file notationCerts[0]=$CA_PEM_FILE \
  --set notation.trustPolicies[0].registryScopes[0]=$REGISTRYSCOPE \
  --set notation.trustPolicies[0].trustStores[0]=ca:notationCerts[0] \
  --set notation.trustPolicies[0].trustedIdentities[0]=$TRUSTEDIDENTITY

export CUSTOM_POLICY=$(curl -L https://raw.githubusercontent.com/deislabs/ratify/main/library/default/customazurepolicy.json)
export DEFINITION_NAME="ratify-default-custom-policy"
export POLICY_SCOPE=$(az aks show -g "${RESOURCE_GROUP}" -n "${CLUSTER_NAME}" --query id -o tsv)

export DEFINITION_ID=$(az policy definition create --name "${DEFINITION_NAME}" --rules "$(echo "${CUSTOM_POLICY}" | jq .policyRule)" --params "$(echo "${CUSTOM_POLICY}" | jq .parameters)" --mode "Microsoft.Kubernetes.Data" --query id -o tsv)

export ASSIGNMENT_ID=$(az policy assignment create --policy "${DEFINITION_ID}" --name "${DEFINITION_NAME}" --scope "${POLICY_SCOPE}" --query id -o tsv)

echo "Please wait policy assignment with id ${ASSIGNMENT_ID} taking effect"
echo "It often requires 15 min"
echo "You can run 'kubectl get constraintTemplate ratifyverification' to verify the policy takes effect"
