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

#!/usr/bin/env bats

load helpers

BATS_TESTS_DIR=${BATS_TESTS_DIR:-test/bats/tests}
WAIT_TIME=60
SLEEP_TIME=1

@test "notation test" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo2 --namespace default --force --ignore-not-found=true'
    }
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    # validate certificate store status property shows success
    run bash -c "kubectl get certificatestores.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n gatekeeper-system -o yaml | grep 'issuccess: true'"
    assert_success
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success
    # validate auth cache and subject descriptor cache hit
    run bash -c "kubectl logs -l app.kubernetes.io/name=ratify -c ratify --tail=-1 -n gatekeeper-system | grep 'auth cache hit'"
    assert_success
    run bash -c "kubectl logs -l app.kubernetes.io/name=ratify -c ratify --tail=-1 -n gatekeeper-system | grep 'cache hit for subject descriptor'"
    assert_success
    sleep 3
    # run a second pod to validate cache hit
    run kubectl run demo2 --namespace default --image=registry:5000/notation:signed
    assert_success
    run bash -c "kubectl logs -l app.kubernetes.io/name=ratify -c ratify --tail=-1 -n gatekeeper-system | grep 'cache hit for subject registry:5000/notation'"
    assert_success
}