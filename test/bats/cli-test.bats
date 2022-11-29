#!/usr/bin/env bats

load helpers

@test "notary verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s libinbinacr.azurecr-test.io/net-monitor:v1
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s libinbinacr.azurecr-test.io/net-monitor:invalid
    assert_cmd_verify_failure
}

@test "cosign verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/cosign-image:signed
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/cosign-image:unsigned
    assert_cmd_verify_failure
}
