#!/usr/bin/env bats

load helpers

@test "notary verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/notation:signed
    assert_cmd_verify_success

    # Test with OCI Artifact Manifest Signature
    if [[ $IS_OCI_1_1 == "true" ]]; then
        run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/notation:ociartifact
        assert_cmd_verify_success
    fi

    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/notation:unsigned
    assert_cmd_verify_failure
}

@test "notation verifier leaf cert test" {
    run bin/ratify verify -c $RATIFY_DIR/config_notation_root_cert.json -s $TEST_REGISTRY/notation:leafSigned
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config_notation_leaf_cert.json -s $TEST_REGISTRY/notation:leafSigned
    assert_cmd_verify_failure
}

@test "cosign verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/cosign:signed-key
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/cosign_keyless_config.json -s wabbitnetworks.azurecr.io/test/cosign-image:signed-keyless
    assert_cmd_verify_success

    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/cosign:unsigned
    assert_cmd_verify_failure
}

@test "licensechecker verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/complete_licensechecker_config.json -s $TEST_REGISTRY/licensechecker:v0
    assert_cmd_verify_success

    # Test with OCI Artifact Manifest Signature
    if [[ $IS_OCI_1_1 == "true" ]]; then
        run bin/ratify verify -c $RATIFY_DIR/complete_licensechecker_config.json -s $TEST_REGISTRY/licensechecker:ociartifact
        assert_cmd_verify_success
    fi

    run bin/ratify verify -c $RATIFY_DIR/partial_licensechecker_config.json -s $TEST_REGISTRY/licensechecker:v0
    assert_cmd_verify_failure
}

@test "sbom verifier test" {
    # Notes: test would fail if sbom/notary types are explicitly specified in the policy
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/sbom:v0
    assert_cmd_verify_success

    # Test with OCI Artifact Manifest Signature
    if [[ $IS_OCI_1_1 == "true" ]]; then
        run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/sbom:ociartifact
        assert_cmd_verify_success
    fi

    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/sbom:unsigned
    assert_cmd_verify_failure
}

@test "schemavalidator verifier test" {
    run bin/ratify verify -c $RATIFY_DIR/schemavalidator_config.json -s $TEST_REGISTRY/schemavalidator:v0
    assert_cmd_verify_success

    # Test with OCI Artifact Manifest Signature
    if [[ $IS_OCI_1_1 == "true" ]]; then
        run bin/ratify verify -c $RATIFY_DIR/schemavalidator_config.json -s $TEST_REGISTRY/schemavalidator:ociartifact
        assert_cmd_verify_success
    fi
}

@test "sbom/notary/cosign/licensechecker verifiers test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s $TEST_REGISTRY/all:v0
    assert_cmd_verify_success
}

@test "dynamic plugin verifier test" {
    # dynamic plugins disabled by default
    run bash -c "bin/ratify verify -c $RATIFY_DIR/dynamic_plugins_config.json -s  $TEST_REGISTRY/all:v0 2>&1 >/dev/null | grep 'dynamic plugins are currently disabled'"
    assert_success

    # dynamic plugins enabled with feature flag
    run bash -c "RATIFY_DYNAMIC_PLUGINS=1 bin/ratify verify -c $RATIFY_DIR/dynamic_plugins_config.json -s  $TEST_REGISTRY/all:v0 2>&1 >/dev/null | grep 'downloaded verifier plugin dynamic from .* to .*'"
    assert_success

    # ensure the plugin is downloaded and marked executable
    test -x $RATIFY_DIR/plugins/dynamic
    assert_success
}

@test "dynamic plugin store test" {
    # dynamic plugins disabled by default
    run bash -c "bin/ratify verify -c $RATIFY_DIR/dynamic_plugins_config.json -s  $TEST_REGISTRY/all:v0 2>&1 >/dev/null | grep 'dynamic plugins are currently disabled'"
    assert_success

    # dynamic plugins enabled with feature flag
    run bash -c "RATIFY_DYNAMIC_PLUGINS=1 bin/ratify verify -c $RATIFY_DIR/dynamic_plugins_config.json -s  $TEST_REGISTRY/all:v0 2>&1 >/dev/null | grep 'downloaded store plugin dynamicstore from .* to .*'"
    assert_success

    # ensure the plugin is downloaded and marked executable
    test -x $RATIFY_DIR/plugins/dynamicstore
    assert_success
}
