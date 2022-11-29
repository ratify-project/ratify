#!/usr/bin/env bats

load helpers

WAIT_TIME=60
SLEEP_TIME=1

@test "ratify verify test" {
    run bin/ratify verify -c $RATIFY_DIR/config.json -s libinbinacr.azurecr-test.io/net-monitor:v1
    echo "actual: $status"
    echo "output: $output"
}
