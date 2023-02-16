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

SUFFIX=$(openssl rand -hex 2)
export GROUP_NAME="${GROUP_NAME:-ratify-e2e-${SUFFIX}}"
export ACR_NAME="${ACR_NAME:-ratifyacr${SUFFIX}}"
export AKS_NAME="${AKS_NAME:-ratify-aks-${SUFFIX}}"
export KEYVAULT_NAME="${KEYVAULT_NAME:-ratify-akv-${SUFFIX}}"
export USER_ASSIGNED_IDENTITY_NAME="${USER_ASSIGNED_IDENTITY_NAME:-ratify-e2e-identity-${SUFFIX}}"
export LOCATION="eastus"
export KUBERNETES_VERSION=${1:-1.24.6}
export RATIFY_NAMESPACE=${4:-default}
export NOTARY_PEM_NAME="notary"
TAG="test${SUFFIX}"
GATEKEEPER_VERSION=${2:-3.11.0}
TENANT_ID=$3
CERT_DIR=${5:-"~/ratify/certs"}

build_push_to_acr() {
  echo "Building and pushing images to ACR"
  docker build --progress=plain --no-cache -f ./httpserver/Dockerfile -t "${ACR_NAME}.azurecr.io/test/localbuild:${TAG}" .
  docker push "${ACR_NAME}.azurecr.io/test/localbuild:${TAG}"

  docker build --progress=plain --no-cache --build-arg KUBE_VERSION=${KUBERNETES_VERSION} --build-arg TARGETOS="linux" --build-arg TARGETARCH="amd64" -f crd.Dockerfile -t "${ACR_NAME}.azurecr.io/test/localbuildcrd:${TAG}" ./charts/ratify/crds
  docker push "${ACR_NAME}.azurecr.io/test/localbuildcrd:${TAG}"
}

deploy_gatekeeper() {
  echo "deploying gatekeeper"
  helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
  helm install gatekeeper/gatekeeper \
    --version ${GATEKEEPER_VERSION} \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=7 \
    --set auditInterval=0
}

deploy_ratify() {
  echo "generating tls certs"
  ./scripts/generate-tls-certs.sh ${CERT_DIR} ${RATIFY_NAMESPACE}

  echo "deploying ratify"
  local IDENTITY_CLIENT_ID=$(az identity show --name ${USER_ASSIGNED_IDENTITY_NAME} --resource-group ${GROUP_NAME} --query 'clientId' -o tsv)
  local VAULT_URI=$(az keyvault show --name ${KEYVAULT_NAME} --resource-group ${GROUP_NAME} --query "properties.vaultUri" -otsv)
  helm install ratify \
    ./charts/ratify --atomic \
    --namespace ${RATIFY_NAMESPACE} --create-namespace \
    --set image.repository=${ACR_NAME}.azurecr.io/test/localbuild \
    --set image.crdRepository=${ACR_NAME}.azurecr.io/test/localbuildcrd \
    --set image.tag=${TAG} \
    --set cosign.enabled=true \
    --set-file cosign.key="./test/testdata/cosign.pub" \
    --set gatekeeper.version=${GATEKEEPER_VERSION} \
    --set-file provider.tls.crt=${CERT_DIR}/server.crt \
    --set-file provider.tls.key=${CERT_DIR}/server.key \
    --set provider.tls.cabundle="$(cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
    --set akvCertConfig.enabled=true \
    --set akvCertConfig.vaultURI=${VAULT_URI} \
    --set akvCertConfig.cert1Name=${NOTARY_PEM_NAME} \
    --set akvCertConfig.tenantId=${TENANT_ID} \
    --set oras.authProviders.azureWorkloadIdentityEnabled=true \
    --set azureWorkloadIdentity.clientId=${IDENTITY_CLIENT_ID}

  kubectl delete verifiers.config.ratify.deislabs.io/verifier-cosign

  kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
  kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
}

save_logs() {
  echo "Saving logs"
  kubectl logs -n gatekeeper-system -l control-plane=controller-manager --tail=-1 >logs-externaldata-controller-aks.json
  kubectl logs -n gatekeeper-system -l control-plane=audit-controller --tail=-1 >logs-externaldata-audit-aks.json
  kubectl logs -n ${RATIFY_NAMESPACE} -l app=ratify --tail=-1 >logs-ratify-preinstall-aks.json
  kubectl logs -n ${RATIFY_NAMESPACE} -l app.kubernetes.io/name=ratify --tail=-1 >logs-ratify-aks.json
}

cleanup() {
  echo "Deleting group"
  az group delete --name "${GROUP_NAME}" --yes --no-wait || true
}

trap cleanup EXIT

main() {
  ./scripts/create-azure-resources.sh

  build_push_to_acr

  deploy_gatekeeper
  deploy_ratify

  bats -t ./test/bats/azure-test.bats
  save_logs
}

main
