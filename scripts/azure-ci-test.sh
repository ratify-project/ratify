#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GROUP_NAME="${GROUP_NAME:-ratify-e2e-$(openssl rand -hex 2)}"
export ACR_NAME="${ACR_NAME:-ratifyacr$(openssl rand -hex 2)}"
export AKS_NAME="${AKS_NAME:-ratify-aks-$(openssl rand -hex 2)}"
export LOCATION="eastus"
KUBERNETES_VERSION=${1:-1.24.6}
GATEKEEPER_VERSION=${2:-3.11.0}
TENANT_ID=$3
RATIFY_NAMESPACE=${4:-default}
CERT_DIR=${5:-"~/ratify/certs"}

build_push_to_acr() {
  echo "Building and pushing images to ACR"
  docker build --progress=plain --no-cache -f ./httpserver/Dockerfile -t "${ACR_NAME}.azurecr.io/test/localbuild:test" .
  docker push "${ACR_NAME}.azurecr.io/test/localbuild:test"

  docker build --progress=plain --no-cache --build-arg KUBE_VERSION=${KUBERNETES_VERSION} --build-arg TARGETOS="linux" --build-arg TARGETARCH="amd64" -f crd.Dockerfile -t "${ACR_NAME}.azurecr.io/test/localbuildcrd:test" ./charts/ratify/crds
  docker push "${ACR_NAME}.azurecr.io/test/localbuildcrd:test"
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
  local IDENTITY_CLIENT_ID=$(az aks show -g ${GROUP_NAME} -n ${AKS_NAME} --query "identityProfile.kubeletidentity.clientId")
  helm install ratify \
    ./charts/ratify --atomic \
    --namespace ${RATIFY_NAMESPACE} --create-namespace \
    --set image.repository=${ACR_NAME}.azurecr.io/test/localbuild \
    --set image.crdRepository=${ACR_NAME}.azurecr.io/test/localbuildcrd \
    --set image.tag=test \
    --set azureManagedIdentity.tenantId=${TENANT_ID} \
    --set oras.authProviders.azureManagedIdentityEnabled=true \
    --set azureManagedIdentity.clientId=${IDENTITY_CLIENT_ID} \
    --set gatekeeper.version=${GATEKEEPER_VERSION} \
    --set-file provider.tls.crt=${CERT_DIR}/server.crt \
    --set-file provider.tls.key=${CERT_DIR}/server.key \
    --set provider.tls.cabundle="$(cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')"

  kubectl apply -f https://deislabs.github.io/ratify/library/default/template.yaml
  kubectl apply -f https://deislabs.github.io/ratify/library/default/samples/constraint.yaml
}

save_logs() {
  echo "Saving logs"
  kubectl logs -n gatekeeper-system -l control-plane=controller-manager --tail=-1 >logs-externaldata-controller-aks.json
  kubectl logs -n gatekeeper-system -l control-plane=audit-controller --tail=-1 >logs-externaldata-audit-aks.json
  kubectl logs -n ratify-service -l app=ratify --tail=-1 >logs-ratify-preinstall-aks.json
  kubectl logs -n ratify-service -l app.kubernetes.io/name=ratify --tail=-1 >logs-ratify-aks.json
}

cleanup() {
  echo "Deleting group"
  az group delete --name "${GROUP_NAME}" --yes --no-wait || true
}

trap cleanup EXIT

main() {
  ./scripts/create-aks-acr.sh

  build_push_to_acr

  deploy_gatekeeper
  deploy_ratify

  bats -t ./test/bats/test.bats

  save_logs
}

main
