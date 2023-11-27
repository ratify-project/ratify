# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash

assert_success() {
  if [[ "$status" != 0 ]]; then
    echo "expected: 0"
    echo "actual: $status"
    echo "output: $output"
    return 1
  fi
}

assert_failure() {
  if [[ "$status" == 0 ]]; then
    echo "expected: non-zero exit code"
    echo "actual: $status"
    echo "output: $output"
    return 1
  fi
}

assert_cmd_verify_success() {
  if [[ "$status" != 0 ]]; then
    return 1
  fi
  if [[ "$output" == *'"isSuccess": false,'* ]]; then
    echo $output
    return 1
  fi
}

assert_cmd_cosign_keyless_verify_bundle_success() {
  if [[ "$status" != 0 ]]; then
    return 1
  fi
  if [[ "$output" == *'"bundleVerified": false,'* ]]; then
    echo $output
    return 1
  fi
}

assert_cmd_verify_failure() {
  if [[ "$status" != 0 ]]; then
    return 1
  fi
  if [[ "$output" == *'"isSuccess": true,'* ]]; then
    echo $output
    return 1
  fi
}

assert_mutate_success() {
  if [[ "$status" != 0 ]]; then
    echo $result
    return 1
  fi
  if [[ "$output" == "" ]]; then
    echo "expected digest to be present in image"
    return 1
  fi
}

wait_for_process() {
  wait_time="$1"
  sleep_time="$2"
  cmd="$3"
  while [ "$wait_time" -gt 0 ]; do
    if eval "$cmd"; then
      return 0
    else
      sleep "$sleep_time"
      echo "# retrying $cmd" >&3
      wait_time=$((wait_time - sleep_time))
    fi
  done
  return 1
}
