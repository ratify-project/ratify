#!/usr/bin/env bash
##--------------------------------------------------------------------
#
# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
##--------------------------------------------------------------------

set -o errexit
set -o nounset
set -o pipefail

: "${AKS_NAME:?AKS_NAME environment variable empty or not defined.}"
: "${ACR_NAME:?ACR_NAME environment variable empty or not defined.}"

register_feature() {
  az provider register --namespace Microsoft.ContainerService
}

main() {
  export -f register_feature
  # might take around 20 minutes to register
  timeout --foreground 1200 bash -c register_feature

  az group create --name "${GROUP_NAME}" --location "${LOCATION}" >/dev/null

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
