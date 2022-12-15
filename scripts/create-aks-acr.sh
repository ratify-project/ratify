#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

: "${AKS_NAME:?Environment variable empty or not defined.}"
: "${ACR_NAME:?Environment variable empty or not defined.}"

register_feature() {
  # pinning to 0.5.87 because of https://github.com/Azure/azure-cli/issues/23267
  az extension add --name aks-preview --version 0.5.87
}

main() {
  export -f register_feature
  # might take around 20 minutes to register
  timeout --foreground 1200 bash -c register_feature
  echo "Creating an AKS cluster '${AKS_NAME}' under group '${GROUP_NAME}'"
  # get the latest patch version of 1.24
  KUBERNETES_VERSION="$(az aks get-versions --location "${LOCATION}" --query 'orchestrators[*].orchestratorVersion' -otsv | grep '1.24' | tail -1)"

  az acr create --name "${ACR_NAME}" \
    --resource-group "${GROUP_NAME}" \
    --sku Standard >/dev/null
  az acr login -n ${ACR_NAME}
  echo "ACR '${ACR_NAME}' is created"

  az aks create \
    --resource-group "${GROUP_NAME}" \
    --name "${AKS_NAME}" \
    --node-vm-size Standard_DS3_v2 \
    --enable-managed-identity \
    --kubernetes-version "${KUBERNETES_VERSION}" \
    --node-count 1 \
    --generate-ssh-keys \
    --attach-acr "${ACR_NAME}" \
    --enable-oidc-issuer >/dev/null

  echo "AKS '${AKS_NAME}' is created and attached to ACR"

  az aks get-credentials --resource-group ${GROUP_NAME} --name ${AKS_NAME}
  echo "Connected to AKS cluster"
}

main
