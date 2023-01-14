#!/usr/bin/env bats

load helpers

@test "notary verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/notary-image:signed
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/notary-image:unsigned
    assert_cmd_verify_failure
}

@test "cosign verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/cosign-image:signed
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/cosign-image:unsigned
    assert_cmd_verify_failure
}

@test "licensechecker verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/complete_licensechecker_config.json -s wabbitnetworks.azurecr.io/test/license-checker-image:v1
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/partial_licensechecker_config.json -s wabbitnetworks.azurecr.io/test/license-checker-image:v1
    assert_cmd_verify_failure
}

@test "sbom verifier test" {
    # Notes: test would fail if sbom/notary types are explicitly specified in the policy
    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/sbom-image:signed
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/sbom-image:unsigned
    assert_cmd_verify_failure
}

@test "schemavalidator verifier test" {   
    # TODO 
    # It's best to use an image with individual artifact types vs an all-in-one so any failures can be isolated.
    # Replace this image reference once we have a local private registry for Ratify.
    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/all-in-one-image:signed
    assert_cmd_verify_success
}

@test "sbom/notary/cosign/licensechecker verifiers test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s wabbitnetworks.azurecr.io/test/all-in-one-image:signed
    assert_cmd_verify_success
}

@test "dynamic plugin verifier test" {
    # dynamic plugins disabled by default
    run bash -c "bin/ratify verify -c $RATIFY_DIR/dynamic_plugins_config.json -s wabbitnetworks.azurecr.io/test/all-in-one-image:signed 2>&1 >/dev/null | grep 'dynamic plugins are currently disabled'"
    assert_success

    # dynamic plugins enabled with feature flag
    run bash -c "RATIFY_DYNAMIC_PLUGINS=1 bin/ratify verify -c $RATIFY_DIR/dynamic_plugins_config.json -s wabbitnetworks.azurecr.io/test/all-in-one-image:signed 2>&1 >/dev/null | grep 'downloaded verifier plugin dynamic from .* to .*'"
    assert_success
}
