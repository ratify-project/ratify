#!/usr/bin/env bats

load helpers
@test "quick start test" {
    # deployment, service and provider for dummy-provider
    run kubectl apply -f https://deislabs.github.io/ratify/charts/ratify-gatekeeper/templates/constraint.yaml
    assert_success
    run kubectl create ns demo
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed -n demo
    assert_success
}