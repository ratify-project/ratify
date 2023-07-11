#!/usr/bin/env bats

load helpers

BATS_TESTS_DIR=${BATS_TESTS_DIR:-test/bats/tests}
WAIT_TIME=60
SLEEP_TIME=1

@test "base test without cert rotator" {
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
    # validate certificate store status property shows success
    run bash -c "kubectl get certificatestores.config.ratify.deislabs.io/ratify-notary-inline-cert -n gatekeeper-system -o yaml | grep 'issuccess: true'"
    assert_success
    run kubectl run demo --namespace default --image=registry:5000/notation:signed
    assert_success

    run kubectl run demo1 --namespace default --image=registry:5000/notation:unsigned
    assert_failure
}

@test "crd version test" {
    run kubectl delete verifiers.config.ratify.deislabs.io/verifier-notary
    assert_success
    run kubectl apply -f ./config/samples/config_v1alpha1_verifier_notary.yaml
    assert_success
    run bash -c "kubectl get verifiers.config.ratify.deislabs.io/verifier-notary -o yaml | grep 'apiVersion: config.ratify.deislabs.io/v1beta1'"
    assert_success

    run kubectl delete stores.config.ratify.deislabs.io/store-oras
    assert_success
    run kubectl apply -f ./config/samples/config_v1alpha1_store_oras_http.yaml
    assert_success
    run bash -c "kubectl get stores.config.ratify.deislabs.io/store-oras -o yaml | grep 'apiVersion: config.ratify.deislabs.io/v1beta1'"
    assert_success
}

@test "notary test" {
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
    # validate certificate store status property shows success
    run bash -c "kubectl get certificatestores.config.ratify.deislabs.io/ratify-notary-inline-cert -n gatekeeper-system -o yaml | grep 'issuccess: true'"
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

    echo "adding license checker, delete notary verifier and validate deployment fails due to missing notary verifier"
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_complete_licensechecker.yaml
    assert_success
    run kubectl delete verifiers.config.ratify.deislabs.io/verifier-notary
    assert_success
    # wait for the httpserver cache to be invalidated
    sleep 15
    run kubectl run crdtest --namespace default --image=registry:5000/notation:signed
    assert_failure

    echo "Add notary verifier and validate deployment succeeds"
    run kubectl apply -f ./config/samples/config_v1beta1_verifier_notary.yaml
    assert_success

    # wait for the httpserver cache to be invalidated
    sleep 15
    run kubectl run crdtest --namespace default --image=registry:5000/notation:signed
    assert_success
}

@test "configmap update test" {
    skip "Skipping test for now as we are no longer watching for configfile update in a k8 environment. This test ensures we are watching config file updates in a non-kub scenario"
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    sleep 5
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    sleep 5
    run kubectl run demo2 --image=registry:5000/notation:signed
    assert_success

    run kubectl get configmaps ratify-configuration --namespace=gatekeeper-system -o yaml >currentConfig.yaml
    run kubectl delete -f ./library/default/samples/constraint.yaml

    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=gatekeeper-system -f ${BATS_TESTS_DIR}/configmap/invalidconfigmap.yaml"
    echo "Waiting for 150 second for configuration update"
    sleep 150

    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success
    run kubectl run demo3 --image=registry:5000/notation:signed
    echo "Current time after validate : $(date +"%T")"
    assert_failure

    wait_for_process ${WAIT_TIME} ${SLEEP_TIME} "kubectl replace --namespace=gatekeeper-system -f currentConfig.yaml"
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

@test "validate inline cert provider" {
    teardown() {
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete certificatestores.config.ratify.deislabs.io/certstore-inline --namespace default --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-alternate --namespace default --force --ignore-not-found=true'

        # restore the original notary verifier for other tests
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f ./config/samples/config_v1beta1_verifier_notary.yaml'
    }

    # configure the default template/constraint
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success

    # verify that the image cannot be run due to an invalid cert
    run kubectl run demo-alternate --namespace default --image=registry:5000/notation:signed-alternate
    assert_failure

    # add the alternate certificate as an inline certificate store
    cat ~/.config/notation/truststore/x509/ca/alternate-cert/alternate-cert.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    assert_success
    sed -i '9,$d' ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml

    # configure the notary verifier to use the inline certificate store
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_verifier_notary.yaml
    assert_success
    sleep 10

    # verify that the image can now be run
    run kubectl run demo-alternate --namespace default --image=registry:5000/notation:signed-alternate
    assert_success
}

@test "validate k8 secrets ORAS auth provider" {
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
    # apply store CRD with k8 secret auth provier enabled
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
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete certificatestores.config.ratify.deislabs.io/certstore-inline --namespace default --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-leaf --namespace default --force --ignore-not-found=true'
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl delete pod demo-leaf2 --namespace default --force --ignore-not-found=true'

        # restore the original notary verifier for other tests
        wait_for_process ${WAIT_TIME} ${SLEEP_TIME} 'kubectl apply -f ./config/samples/config_v1beta1_verifier_notary.yaml'
    }

    # configure the default template/constraint
    run kubectl apply -f ./library/default/template.yaml
    assert_success
    run kubectl apply -f ./library/default/samples/constraint.yaml
    assert_success

    # add the root certificate as an inline certificate store
    cat ~/.config/notation/truststore/x509/ca/leaf-test/root.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    assert_success
    sed -i '9,$d' ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml

    # configure the notary verifier to use the inline certificate store
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_verifier_notary.yaml
    assert_success

    # verify that the image can be run with a root cert
    run kubectl run demo-leaf --namespace default --image=registry:5000/notation:leafSigned
    assert_success

    # add the root certificate as an inline certificate store
    cat ~/.config/notation/truststore/x509/ca/leaf-test/leaf.crt | sed 's/^/      /g' >>./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    run kubectl apply -f ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml
    assert_success
    sed -i '9,$d' ./test/bats/tests/config/config_v1beta1_certstore_inline.yaml

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
    run kubectl get secret ratify-tls -n gatekeeper-system -o json | jq --arg cert "$(cat .staging/rotation/server.crt | base64)" --arg key "$(cat .staging/rotation/server.key | base64)" '.data["tls.key"]=$key | .data["tls.crt"]=$cert' | kubectl replace -f -

    # update the gatekeeper webhook server tls secret to use the new cert bundle
    run kubectl get Secret gatekeeper-webhook-server-cert -n gatekeeper-system -o json | jq --arg caCert "$(cat .staging/rotation/gatekeeper/ca.crt | base64)" --arg caKey "$(cat .staging/rotation/gatekeeper/ca.key | base64)" --arg tlsCert "$(cat .staging/rotation/gatekeeper/server.crt | base64)" --arg tlsKey "$(cat .staging/rotation/gatekeeper/server.key | base64)" '.data["ca.crt"]=$caCert | .data["ca.key"]=$caKey | .data["tls.crt"]=$tlsCert | .data["tls.key"]=$tlsKey' | kubectl replace -f -

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
