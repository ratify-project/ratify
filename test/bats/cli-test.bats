#!/usr/bin/env bats

load helpers

WAIT_TIME=60
SLEEP_TIME=1

@test "notation verify" {
    notation version
}
