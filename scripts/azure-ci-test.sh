#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GROUP_NAME="ratify-e2e"
export LOCATION="eastus"
export ACR_NAME="${ACR_NAME:-ratifyacr$(openssl rand -hex 2)}"
export AKS_NAME="${AKS_NAME:-ratify-e2e-$(openssl rand -hex 2)}"

cluster_exists() {
  if az aks show --resource-group "${GROUP_NAME}" --name "$1" >/dev/null; then
    echo "true" && return
  fi
  echo "false" && return
}

acr_exists() {
  count=$(az acr list -g ratify-e2e -o table | grep "$1" | wc -l)
  if [[ $count -eq 1 ]]; then
    echo "true" && return
  fi
  echo "false" && return
}

get_new_cluster_name() {
  name="ratify-e2e-$(openssl rand -hex 2)"
  while [[ "$(cluster_exists $name)" == "true" ]]; do
    name="ratify-e2e-$(openssl rand -hex 2)"
  done
  export AKS_NAME=$name
}

get_new_acr_name() {
  name="ratifyacr$(openssl rand -hex 2)"
  while [[ "$(acr_exists $name)" == "true" ]]; do
    name="ratifyacr$(openssl rand -hex 2)"
  done
  export ACR_NAME=$name
}

build_push_to_acr() {
  echo "Building and pushing images to ACR"
  docker build --progress=plain --no-cache -f ./httpserver/Dockerfile -t "${ACR_NAME}.azurecr.io/test/localbuild:test" .
  docker push "${ACR_NAME}.azurecr.io/test/localbuild:test"

  docker build --progress=plain --no-cache --build-arg KUBE_VERSION=$1 --build-arg TARGETOS="linux" --build-arg TARGETARCH="amd64" -f crd.Dockerfile -t "${ACR_NAME}.azurecr.io/test/localbuildcrd:test" ./charts/ratify/crds
  docker push "${ACR_NAME}.azurecr.io/test/localbuildcrd:test"
}

deploy_gatekeeper() {
  echo "deploying gatekeeper"
  helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
  helm install gatekeeper/gatekeeper \
    --version 3.10.0 \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=7 \
    --set auditInterval=0
}

deploy_ratify() {
  echo "deploying ratify"
  local IDENTITY_CLIENT_ID=$(az aks show -g ${GROUP_NAME} -n ${AKS_NAME} --query "identityProfile.kubeletidentity.clientId")
  helm install ratify \
    ./charts/ratify --atomic \
    --set image.repository=${ACR_NAME}.azurecr.io/test/localbuild \
    --set image.crdRepository=${ACR_NAME}.azurecr.io/test/localbuildcrd \
    --set image.tag=test \
    --set azureManagedIdentity.tenantId=${TENANT_ID} \
    --set oras.authProviders.azureManagedIdentityEnabled=true \
    --set azureManagedIdentity.clientId=${IDENTITY_CLIENT_ID} \
    --namespace ratify-service --create-namespace

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
  echo "Deleting ACR and AKS resources"
  az aks delete -n "${AKS_NAME}" -g "${GROUP_NAME}" --yes --no-wait || true
  az acr delete -n "${ACR_NAME}" -g "${GROUP_NAME}" --yes || true
}

trap cleanup EXIT

main() {
  local KUBERNETES_VERSION=$1
  export TENANT_ID=$2

  get_new_cluster_name
  get_new_acr_name

  ./scripts/create-aks-acr.sh

  build_push_to_acr $KUBERNETES_VERSION

  deploy_gatekeeper
  deploy_ratify

  bats -t ./test/bats/test.bats

  save_logs
}

main $1 $2
