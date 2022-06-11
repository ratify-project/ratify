#!/usr/bin/env bats

load helpers
@test "quick start test" {
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    run kubectl run demo --image=ratify.azurecr.io/testimage:signed
    assert_success
    run kubectl run demo1 --image=ratify.azurecr.io/testimage:unsigned
    assert_failure
}
