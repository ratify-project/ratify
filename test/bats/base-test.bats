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
RATIFY_NAMESPACE=gatekeeper-system

@test "base test without cert rotator" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo1 --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod initcontainer-pod --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod initcontainer-pod1 --namespace default --force --ignore-not-found=true'
    }
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    # validate key management provider status property shows success
    run bash -c "kubectl get keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE} -o yaml | grep 'issuccess: true'"
    assert_success
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success
    run kubectl run demo1 --namespace default --image=registry:5000/notation:unsigned
    assert_failure

    # validate initContainers image
    run kubectl apply -f ./test/testdata/pod_initContainers_signed.yaml --namespace default
    assert_success
    
    run kubectl apply -f ./test/testdata/pod_initContainers_unsigned.yaml --namespace default
    assert_failure

    # validate ephemeralContainers image
    run kubectl debug demo --image=registry:5000/notation:signed --target=demo
    assert_success

    run kubectl debug demo --image=registry:5000/notation:unsigned --target=demo
    assert_failure
}

@test "crd version test" {
    run kubectl delete verifiers.config.ratify.deislabs.io/verifier-notation
    assert_success
    run kubectl apply -f ./config/samples/config_v1alpha1_verifier_notation.yaml
    assert_success
    run bash -c "kubectl get verifiers.config.ratify.deislabs.io/verifier-notation -o yaml | grep 'apiVersion: config.ratify.deislabs.io/v1beta1'"
    assert_success

    run kubectl delete stores.config.ratify.deislabs.io/store-oras
    assert_success
    run kubectl apply -f ./config/samples/config_v1alpha1_store_oras_http.yaml
    assert_success
    run bash -c "kubectl get stores.config.ratify.deislabs.io/store-oras -o yaml | grep 'apiVersion: config.ratify.deislabs.io/v1beta1'"
    assert_success
}

@test "notation test" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo1 --namespace default --force --ignore-not-found=true'
    }
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5

    # validate key management provider status property shows success
    run bash -c "kubectl get keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE} -o yaml | grep 'issuccess: true'"
    assert_success
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success

    run kubectl run demo1 --namespace default --image=registry:5000/notation:unsigned
    assert_failure
}

@test "notation test with certs across namespace" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo1 --namespace default --force --ignore-not-found=true'
        
        # restore cert store in ratify namespace
        run bash -c "kubectl get keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -o yaml -n default > kmprovider.yaml"
        run kubectl delete keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n default
        sed 's/default/gatekeeper-system/' kmprovider.yaml > kmproviderNewNS.yaml
        run kubectl apply -f kmproviderNewNS.yaml        
        assert_success

        # restore the original notation verifier for other tests
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f ./config/samples/config_v1beta1_verifier_notation.yaml'
    }
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml

    assert_success
    sleep 5
    
    # apply the key management provider to default namespace
    run bash -c "kubectl get keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -o yaml -n ${RATIFY_NAMESPACE} > kmprovider.yaml"    
    assert_success
    sed 's/gatekeeper-system/default/' kmprovider.yaml > kmproviderNewNS.yaml    
    run kubectl apply -f kmproviderNewNS.yaml
    assert_success
    run kubectl delete keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE}
    assert_success
    
    # configure the notation verifier to use inline certificate store with specific namespace
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_notation_specificnskmprovider.yaml
    assert_success

    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success

    run kubectl run demo1 --namespace default --image=registry:5000/notation:unsigned
    assert_failure
}

@test "validate crd add, replace and delete" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod crdtest --namespace default --force --ignore-not-found=true'
    }

    echo "adding license checker, delete notation verifier and validate deployment fails due to missing notation verifier"
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_complete_licensechecker.yaml
    assert_success
    run kubectl delete verifiers.config.ratify.deislabs.io/verifier-notation
    assert_success
    # wait for the httpserver cache to be invalidated
    sleep 15
    run kubectl run crdtest --namespace default --image=registry:5000/notation:signed
    assert_failure

    echo "Add notation verifier and validate deployment succeeds"
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_notation.yaml
    assert_success

    # wait for the httpserver cache to be invalidated
    sleep 15
    run kubectl run crdtest --namespace default --image=registry:5000/notation:signed
    assert_success
}

@test "verifier crd status check" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete verifiers.config.ratify.deislabs.io/verifier-license-checker'
    }

    # apply a valid verifier, validate status property shows success
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_complete_licensechecker.yaml
    assert_success
    run bash -c "kubectl describe verifiers.config.ratify.deislabs.io/verifier-license-checker -n ${RATIFY_NAMESPACE} | grep 'Issuccess:  true'"
    assert_success

    # apply a invalid verifier CR, validate status with error
    sed 's/licensechecker/invalidlicensechecker/' ./config/samples/config_v1beta1_verifier_complete_licensechecker.yaml > invalidVerifier.yaml
    run kubectl apply -f invalidVerifier.yaml
    assert_success
    run bash -c "kubectl describe verifiers.config.ratify.deislabs.io/verifier-license-checker -n ${RATIFY_NAMESPACE} | grep 'Brieferror:  Original Error:'"
    assert_success

    # apply a valid verifier, validate status property shows success
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_complete_licensechecker.yaml
    assert_success
    run bash -c "kubectl describe verifiers.config.ratify.deislabs.io/verifier-license-checker -n ${RATIFY_NAMESPACE} | grep 'Issuccess:  true'"
    assert_success
    run bash -c "kubectl describe verifiers.config.ratify.deislabs.io/verifier-license-checker -n ${RATIFY_NAMESPACE} | grep 'Brieferror:  Original Error:'"
    assert_failure
}

@test "store crd status check" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete stores.config.ratify.deislabs.io/store-dynamic'
    }

    # apply a invalid verifier CR, validate status with error
    sed 's/:v1/:invalid/' ./config/samples/config_v1beta1_store_dynamic.yaml > invalidstore.yaml
    run kubectl apply -f invalidstore.yaml
    assert_success
    # wait for download of image
    sleep 5
    run bash -c "kubectl describe stores.config.ratify.deislabs.io/store-dynamic -n ${RATIFY_NAMESPACE} | grep 'plugin not found'"
    assert_success
}

@test "configmap update test" {
    skip "Skipping test for now as we are no longer watching for configfile update in a K8s environment. This test ensures we are watching config file updates in a non-kub scenario"
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run demo2 --image=registry:5000/notation:signed
    assert_success

    run kubectl get configmaps ratify-configuration --namespace=${RATIFY_NAMESPACE} -o yaml >currentConfig.yaml
    run kubectl delete -f ./library/default/samples/constraint.yaml

    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=${RATIFY_NAMESPACE} -f ${BATS_TESTS_DIR}/configmap/invalidconfigmap.yaml"
    echo "Waiting for 150 second for configuration update"
    sleep 150

    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    run kubectl run demo3 --image=registry:5000/notation:signed
    echo "Current time after validate : $(date +"%T")"
    assert_failure

    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=${RATIFY_NAMESPACE} -f currentConfig.yaml"
}

@test "validate mutation tag to digest" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod mutate-demo --namespace default --ignore-not-found=true'
    }

    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run mutate-demo --namespace default --image=registry:5000/notation:signed
    assert_success
    result=$(kubectl get pod mutate-demo --namespace default -o json | jq -r ".spec.containers[0].image" | grep @sha)
    assert_mutate_success
}

@test "validate inline certificate store provider" {
    teardown() {
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete certificatestores.config.ratify.deislabs.io/certstore-inline --namespace ${RATIFY_NAMESPACE} --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-alternate --namespace default --force --ignore-not-found=true'

        # restore the original key management provider
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f kmprovider_staging.yaml'
        # restore the original notation verifier for other tests
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f ./config/samples/config_v1beta1_verifier_notation.yaml'
    }

    # save the existing key management provider inline resource to restore later
    run bash -c "kubectl get keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE} -o yaml > kmprovider_staging.yaml"
    assert_success
    # configure the default template/constraint
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success

    # verify that the image cannot be run due to an invalid cert
    run kubectl run demo-alternate --namespace default --image=registry:5000/notation:signed-alternate
    assert_failure

    # delete the existing key management provider inline resource since certificate store and key management provider cannot be used together
    run kubectl delete keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE}
    assert_success
    # add the alternate certificate as an inline certificate store
    cat ~/.config/notation/truststore/x509/ca/alternate-cert/alternate-cert.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml --namespace ${RATIFY_NAMESPACE}
    assert_success
    sed -i '9,$d' ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml

    # configure the notation verifier to use the inline certificate store
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_verifier_notation.yaml
    assert_success
    sleep 10

    # verify that the image can now be run
    run kubectl run demo-alternate --namespace default --image=registry:5000/notation:signed-alternate
    assert_success
}

@test "validate inline key management provider" {
    teardown() {
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete keymanagementproviders.config.ratify.deislabs.io/keymanagementprovider-inline --namespace ${RATIFY_NAMESPACE} --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-alternate --namespace default --force --ignore-not-found=true'

        # restore the original notation verifier for other tests
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f ./config/samples/config_v1beta1_verifier_notation.yaml'
    }

    # configure the default template/constraint
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success

    # verify that the image cannot be run due to an invalid cert
    sleep 10
    run kubectl run demo-alternate --namespace default --image=registry:5000/notation:signed-alternate
    assert_failure
    sleep 10

    # add the alternate certificate as an inline key management provider
    cat ~/.config/notation/truststore/x509/ca/alternate-cert/alternate-cert.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml --namespace ${RATIFY_NAMESPACE}
    assert_success
    sed -i '10,$d' ./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml

    # configure the notation verifier to use the inline key management provider
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_verifier_notation_kmprovider.yaml
    assert_success
    sleep 10

    # verify that the image can now be run
    run kubectl run demo-alternate --namespace default --image=registry:5000/notation:signed-alternate
    assert_success
}

@test "validate inline key management provider with inline certificate store" {
    # this test validates that if a key management provider and certificate store are both configured with the same name,
    # the key management provider will take precedence and continue to work as expected
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo1 --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete certificatestores.config.ratify.deislabs.io/ratify-notation-inline-cert-0 --namespace ${RATIFY_NAMESPACE} --ignore-not-found=true'
    }
     # configure the default template/constraint
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success

    # validate key management provider status property shows success
    run bash -c "kubectl get keymanagementproviders.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE} -o yaml | grep 'issuccess: true'"
    assert_success
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success

    sleep 10

    # apply an invalid cert in an inline certificate store 
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_certstore_inline_invalid.yaml -n ${RATIFY_NAMESPACE}
    assert_success
    # validate key management provider status property shows success
    run bash -c "kubectl get certificatestores.config.ratify.deislabs.io/ratify-notation-inline-cert-0 -n ${RATIFY_NAMESPACE} -o yaml | grep 'issuccess: true'"
    assert_success
    # verification should succeed as the existing KMP will take precedence over the new certificate store
    run kubectl run demo1 --namespace default --image=registry:5000/notation:signed
    assert_success

}

@test "validate K8s secrets ORAS auth provider" {
    teardown() {
        echo "cleaning up"
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo1 --namespace default --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl replace -f ./config/samples/config_v1beta1_store_oras_http.yaml'
    }

    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    # apply store CRD with K8s secret auth provier enabled
    run kubectl apply -f ./config/samples/config_v1beta1_store_oras_k8secretAuth.yaml
    assert_success
    sleep 5
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success
    run kubectl run demo1 --namespace default --image=registry:5000/notation:unsigned
    assert_failure
}

@test "validate image signed by leaf cert" {
    teardown() {
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete keymanagementproviders.config.ratify.deislabs.io/keymanagementprovider-inline --namespace ${RATIFY_NAMESPACE} --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-leaf --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-leaf2 --namespace default --force --ignore-not-found=true'

        # restore the original notation verifier for other tests
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f ./config/samples/config_v1beta1_verifier_notation.yaml'
    }

    # configure the default template/constraint
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success

    # add the root certificate as an inline key management provider
    cat ~/.config/notation/truststore/x509/ca/leaf-test/root.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml --namespace ${RATIFY_NAMESPACE}
    assert_success
    sed -i '10,$d' ./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml

    # configure the notation verifier to use the inline key management provider
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_verifier_notation_kmprovider.yaml
    assert_success

    # verify that the image can be run with a root cert
    run kubectl run demo-leaf --namespace default --image=registry:5000/notation:leafSigned
    assert_success

    # add the root certificate as an inline key management provider
    cat ~/.config/notation/truststore/x509/ca/leaf-test/leaf.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml --namespace ${RATIFY_NAMESPACE}
    assert_success
    sed -i '10,$d' ./test/bats/tests/config/config_v1beta1_keymanagementprovider_inline.yaml

    # wait for the httpserver cache to be invalidated
    sleep 15
    # verify that the image cannot be run with a leaf cert
    run kubectl run demo-leaf2 --namespace default --image=registry:5000/notation:leafSigned
    assert_failure
}

@test "validate ratify/gatekeeper tls cert rotation" {
    teardown() {
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo --namespace default --force --ignore-not-found=true'
    }

    # update Providers to use the new CA
    run kubectl get Provider ratify-mutation-provider -o json | jq --arg ca "$(cat .staging/rotation/ca.crt | base64)" '.spec.caBundle=$ca' | kubectl replace -f -
    run kubectl get Provider ratify-provider -o json | jq --arg ca "$(cat .staging/rotation/ca.crt | base64)" '.spec.caBundle=$ca' | kubectl replace -f -

    # update the ratify tls secret to use the new tls cert and key
    run kubectl get secret ratify-tls -n ${RATIFY_NAMESPACE} -o json | jq --arg cert "$(cat .staging/rotation/server.crt | base64)" --arg key "$(cat .staging/rotation/server.key | base64)" '.data["tls.key"]=$key | .data["tls.crt"]=$cert' | kubectl replace -f -

    # update the gatekeeper webhook server tls secret to use the new cert bundle
    run kubectl get Secret gatekeeper-webhook-server-cert -n ${RATIFY_NAMESPACE} -o json | jq --arg caCert "$(cat .staging/rotation/gatekeeper/ca.crt | base64)" --arg caKey "$(cat .staging/rotation/gatekeeper/ca.key | base64)" --arg tlsCert "$(cat .staging/rotation/gatekeeper/server.crt | base64)" --arg tlsKey "$(cat .staging/rotation/gatekeeper/server.key | base64)" '.data["ca.crt"]=$caCert | .data["ca.key"]=$caKey | .data["tls.crt"]=$tlsCert | .data["tls.key"]=$tlsKey' | kubectl replace -f -

    # volume projection can take up to 90 seconds
    sleep 100

    # verify that the verification succeeds
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success
}
