#!/usr/bin/env bats

load helpers

BATS_TESTS_DIR=${BATS_TESTS_DIR:-test/bats/tests}
WAIT_TIME=60
SLEEP_TIME=1

@test "base test without cert rotator" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
    }
    run kubectl run demo --image=ghcr.io/deislabs/ratify/notary-image:signed
    assert_success
 
    # validate unsigned fails
    kubectl run demo1 --image=ghcr.io/deislabs/ratify/notary-image:unsigned
    assert_failure
}
