#!/usr/bin/env bats

load helpers

BATS_TESTS_DIR=${BATS_TESTS_DIR:-test/bats/tests}
WAIT_TIME=60
SLEEP_TIME=1

@test "validate quick start steps" {
    run kubectl run demo --image=ghcr.io/deislabs/ratify/notary-image:signed
    assert_success
 
    # validate unsigned fails
    run kubectl run demo1 --image=ghcr.io/deislabs/ratify/notary-image:unsigned
    assert_failure
}
