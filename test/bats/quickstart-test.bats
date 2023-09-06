#!/usr/bin/env bats

load helpers

@test "validate quick start steps" {
    run kubectl run demo --image=ghcr.io/deislabs/ratify/notary-image:signed
    assert_success
 
    # validate unsigned fails
    run kubectl run demo1 --image=ghcr.io/deislabs/ratify/notary-image:unsigned
    assert_failure
}