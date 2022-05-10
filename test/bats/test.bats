#!/usr/bin/env bats

load helpers
@test "quick start test" {
    # deployment, service and provider for dummy-provider
    run kubectl apply -f ./charts/ratify-gatekeeper/templates/constraint.yaml
    assert_success
    run kubectl create ns demo
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed -n demo
    assert_success
    run kubectl run demo1 --image=ratify.azurecr.io/testimage:unsigned -n demo
    assert_failure
    run kubectl delete namespace demo
}