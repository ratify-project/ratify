#!/usr/bin/env bats

load helpers

BATS_TESTS_DIR=${BATS_TESTS_DIR:-test/bats/tests}
WAIT_TIME=120
SLEEP_TIME=5
CLEAN_CMD="echo cleaning..."

teardown() {
  bash -c "${CLEAN_CMD}"
}

@test "quick start test" {
    run kubectl apply -f ./charts/ratify-gatekeeper/templates/constraint.yaml
    assert_success
    run kubectl create ns demo
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed -n demo
    assert_success
    run kubectl run demo1 --image=ratify.azurecr.io/testimage:unsigned -n demo
    assert_failure
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl delete namespace demo"
}

@test "configmap update test" {
    run kubectl apply -f ./charts/ratify-gatekeeper/templates/constraint.yaml
    run kubectl create ns demo
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed -n demo
    assert_success

    run kubectl get configmaps ratify-configuration --namespace=ratify-service -o yaml > currentConfig.yaml
    run kubectl delete -f https://deislabs.github.io/ratify/charts/ratify-gatekeeper/templates/constraint.yaml
    run kubectl delete namespace demo
                                            
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=ratify-service -f ${BATS_TESTS_DIR}/configmap/invalidconfigmap.yaml"
    echo "Current time after replace1 : $(date +"%T")"
    
    echo "Current time after sleep : $(date +"%T")"
    run kubectl apply -f ./charts/ratify-gatekeeper/templates/constraint.yaml
    run kubectl create ns demo
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed -n demo
    echo "Current time after validate : $(date +"%T")"
     
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=ratify-service -f currentConfig.yaml"
    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl delete namespace demo"
}
