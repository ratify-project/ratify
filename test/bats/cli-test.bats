#!/usr/bin/env bats

load helpers

@test "notary verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/notation:signed
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/notation:unsigned
    assert_cmd_verify_failure
}

@test "cosign verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/cosign:signed
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/cosign:unsigned
    assert_cmd_verify_failure
}

@test "licensechecker verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/complete_licensechecker_config.json -s $LOCAL_TEST_REGISTRY/licensechecker:v0
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/partial_licensechecker_config.json -s $LOCAL_TEST_REGISTRY/licensechecker:v0
    assert_cmd_verify_failure
}

@test "sbom verifier test" {
    # Notes: test would fail if sbom/notary types are explicitly specified in the policy
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/sbom:v0
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/sbom:unsigned
    assert_cmd_verify_failure
}

@test "schemavalidator verifier test" {   
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/schemavalidator:v0
    assert_cmd_verify_success
}

@test "sbom/notary/cosign/licensechecker verifiers test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $LOCAL_TEST_REGISTRY/all:v0
    assert_cmd_verify_success
}
