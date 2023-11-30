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
	  
BINARY_NAME		= ratify
INSTALL_DIR		= ~/.ratify
CERT_DIR        = ${GITHUB_WORKSPACE}/tls/certs

GO_PKG			= github.com/deislabs/ratify
GIT_COMMIT_HASH = $(shell git rev-parse HEAD)
GIT_TREE_STATE 	= $(shell test -n "`git status --porcelain`" && echo "modified" || echo "unmodified")
GIT_TAG     	= $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

LDFLAGS = -w
LDFLAGS += -X $(GO_PKG)/internal/version.GitCommitHash=$(GIT_COMMIT_HASH)
LDFLAGS += -X $(GO_PKG)/internal/version.GitTreeState=$(GIT_TREE_STATE)
LDFLAGS += -X $(GO_PKG)/internal/version.GitTag=$(GIT_TAG)

KIND_VERSION ?= 0.14.0
KUBERNETES_VERSION ?= 1.27.7
KIND_KUBERNETES_VERSION ?= 1.27.3
GATEKEEPER_VERSION ?= 3.14.0
DAPR_VERSION ?= 1.11.1
COSIGN_VERSION ?= 1.13.1
NOTATION_VERSION ?= 1.0.0-rc.7
ORAS_VERSION ?= 1.0.0-rc.2

HELM_VERSION ?= 3.9.2
HELMFILE_VERSION ?= 0.155.0
BATS_BASE_TESTS_FILE ?= test/bats/base-test.bats
BATS_PLUGIN_TESTS_FILE ?= test/bats/plugin-test.bats
BATS_CLI_TESTS_FILE ?= test/bats/cli-test.bats
BATS_QUICKSTART_TESTS_FILE ?= test/bats/quickstart-test.bats
BATS_HA_TESTS_FILE ?= test/bats/high-availability.bats
BATS_VERSION ?= 1.7.0
SYFT_VERSION ?= v0.76.0
YQ_VERSION ?= v4.34.1
YQ_BINARY ?= yq_linux_amd64
ALPINE_IMAGE ?= alpine@sha256:93d5a28ff72d288d69b5997b8ba47396d2cbb62a72b5d87cd3351094b5d578a0
REDIS_IMAGE_TAG ?= 7.0-debian-11
CERT_ROTATION_ENABLED ?= false
REGO_POLICY_ENABLED ?= false
SBOM_TOOL_VERSION ?=v1.2.0

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.2

GATEKEEPER_NAMESPACE = gatekeeper-system
RATIFY_NAME = ratify

# Local Registry Setup
LOCAL_REGISTRY_IMAGE ?= ghcr.io/oras-project/registry:v1.0.0-rc.4
LOCAL_UNSIGNED_IMAGE = hello-world:latest
TEST_REGISTRY = localhost:5000
TEST_REGISTRY_USERNAME = test_user
TEST_REGISTRY_PASSWORD = test_pw

all: build test

.PHONY: build
build: build-cli build-plugins

.PHONY: build-cli
build-cli: fmt vet
	go build --ldflags="$(LDFLAGS)" -cover \
	-coverpkg=github.com/deislabs/ratify/pkg/...,github.com/deislabs/ratify/config/...,github.com/deislabs/ratify/cmd/... \
	-o ./bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

.PHONY: build-plugins
build-plugins:
	go build -cover -coverpkg=github.com/deislabs/ratify/plugins/verifier/cosign/... -o ./bin/plugins/ ./plugins/verifier/cosign
	go build -cover -coverpkg=github.com/deislabs/ratify/plugins/verifier/licensechecker/... -o ./bin/plugins/ ./plugins/verifier/licensechecker
	go build -cover -coverpkg=github.com/deislabs/ratify/plugins/verifier/sample/... -o ./bin/plugins/ ./plugins/verifier/sample
	go build -cover -coverpkg=github.com/deislabs/ratify/plugins/verifier/sbom/... -o ./bin/plugins/ ./plugins/verifier/sbom
	go build -cover -coverpkg=github.com/deislabs/ratify/plugins/verifier/schemavalidator/... -o ./bin/plugins/ ./plugins/verifier/schemavalidator

.PHONY: install
install:
	mkdir -p ${INSTALL_DIR}
	mkdir -p ${INSTALL_DIR}/ratify-certs/cosign
	mkdir -p ${INSTALL_DIR}/ratify-certs/notation
	cp -r ./bin/* ${INSTALL_DIR}

.PHONY: ratify-config
ratify-config:
	cp ./test/bats/tests/config/* ${INSTALL_DIR}
	cp ./test/bats/tests/certificates/wabbit-networks.io.crt ${INSTALL_DIR}/ratify-certs/notation/wabbit-networks.io.crt
	cp ./test/bats/tests/certificates/cosign.pub ${INSTALL_DIR}/ratify-certs/cosign/cosign.pub
	cp -r ./test/bats/tests/schemas/ ${INSTALL_DIR}
	
.PHONY: test
test:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: clean
clean:
	go clean
	rm ./bin/${BINARY_NAME}

.PHONY: deploy-demo
deploy-demo: deploy-gatekeeper deploy-ratify deploy-demo-constraints

.PHONY: delete-demo
delete-demo: delete-demo-constraints delete-ratify delete-gatekeeper

.PHONY: deploy-ratify
deploy-ratify:
	helm install ratify ./charts/ratify --atomic

.PHONY: delete-ratify
delete-ratify:
	helm delete ratify

.PHONY: deploy-demo-constraints
deploy-demo-constraints:
	kubectl apply -f ./library/default/template.yaml
	kubectl apply -f ./library/default/samples/constraint.yaml

.PHONY: delete-demo-constraints
delete-demo-constraints:
	kubectl delete -f ./library/default/template.yaml
	kubectl delete -f ./library/default/samples/constraint.yaml

.PHONY: deploy-rego-policy
deploy-rego-policy:
	kubectl apply -f ./config/samples/policy/config_v1beta1_policy_rego.yaml

.PHONY: deploy-gatekeeper
deploy-gatekeeper:
	helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
	helm install gatekeeper/gatekeeper  \
		--version ${GATEKEEPER_VERSION} \
		--name-template=gatekeeper \
		--namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
	    --set enableExternalData=true

.PHONY: delete-gatekeeper
delete-gatekeeper:
	helm delete gatekeeper --namespace ${GATEKEEPER_NAMESPACE}

.PHONY: test-e2e
test-e2e: generate-rotation-certs
	timeout 20m bats -t ${BATS_BASE_TESTS_FILE}
	EXPIRING_CERT_DIR=.staging/rotation/expiring-certs CERT_DIR=.staging/rotation GATEKEEPER_VERSION=${GATEKEEPER_VERSION} bats -t ${BATS_PLUGIN_TESTS_FILE}

.PHONY: test-e2e-cli
test-e2e-cli: e2e-dependencies e2e-create-local-registry e2e-notation-setup e2e-notation-leaf-cert-setup e2e-cosign-setup e2e-licensechecker-setup e2e-sbom-setup e2e-schemavalidator-setup
	rm ${GOCOVERDIR} -rf
	mkdir ${GOCOVERDIR} -p
	RATIFY_DIR=${INSTALL_DIR} TEST_REGISTRY=${TEST_REGISTRY} ${GITHUB_WORKSPACE}/bin/bats -t ${BATS_CLI_TESTS_FILE}
	go tool covdata textfmt -i=${GOCOVERDIR} -o test/e2e/coverage.txt

.PHONY: test-quick-start
test-quick-start:
	bats -t ${BATS_QUICKSTART_TESTS_FILE}

.PHONY: test-high-availability
test-high-availability:
	bats -t ${BATS_HA_TESTS_FILE}

.PHONY: generate-certs
generate-certs:
	./scripts/generate-tls-certs.sh ${CERT_DIR} ${GATEKEEPER_NAMESPACE}

generate-rotation-certs:
	mkdir -p .staging/rotation
	mkdir -p .staging/rotation/gatekeeper
	mkdir -p .staging/rotation/expiring-certs

	./scripts/generate-gk-tls-certs.sh .staging/rotation/gatekeeper ${GATEKEEPER_NAMESPACE}
	./scripts/generate-tls-certs.sh .staging/rotation ${GATEKEEPER_NAMESPACE}
	./scripts/generate-tls-certs.sh .staging/rotation/expiring-certs ${GATEKEEPER_NAMESPACE} 1

install-bats:
	# Download and install bats
	curl -sSLO https://github.com/bats-core/bats-core/archive/v${BATS_VERSION}.tar.gz && tar -zxvf v${BATS_VERSION}.tar.gz && bash bats-core-${BATS_VERSION}/install.sh ${GITHUB_WORKSPACE}

.PHONY: lint
lint:
	# Download and install golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# Run golangci-lint
	golangci-lint run --print-issued-lines=false --out-format=colored-line-number --issues-exit-code=0

e2e-dependencies:
	# Download and install kind
	curl -L https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-linux-amd64 --output ${GITHUB_WORKSPACE}/bin/kind && chmod +x ${GITHUB_WORKSPACE}/bin/kind
	# Download and install kubectl
	curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl --output ${GITHUB_WORKSPACE}/bin/kubectl && chmod +x ${GITHUB_WORKSPACE}/bin/kubectl
	# Download and install bats
	curl -sSLO https://github.com/bats-core/bats-core/archive/v${BATS_VERSION}.tar.gz && tar -zxvf v${BATS_VERSION}.tar.gz && bash bats-core-${BATS_VERSION}/install.sh ${GITHUB_WORKSPACE}
	# Download and install jq
	curl -L https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 --output ${GITHUB_WORKSPACE}/bin/jq && chmod +x ${GITHUB_WORKSPACE}/bin/jq
	# Download and install yq
	curl -L https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/${YQ_BINARY} --output ${GITHUB_WORKSPACE}/bin/yq && chmod +x ${GITHUB_WORKSPACE}/bin/yq
	# Install ORAS
	curl -LO https://github.com/oras-project/oras/releases/download/v${ORAS_VERSION}/oras_${ORAS_VERSION}_linux_amd64.tar.gz
	mkdir -p oras-install/
	tar -zxf oras*.tar.gz -C oras-install/
	mv oras-install/oras ${GITHUB_WORKSPACE}/bin
	rm -rf oras*.tar.gz oras-install/

KIND_NODE_VERSION := kindest/node:v$(KIND_KUBERNETES_VERSION)

e2e-create-local-registry: e2e-run-local-registry e2e-create-all-image

e2e-run-local-registry:
	rm -rf ~/auth
	mkdir ~/auth
	docker run --entrypoint htpasswd httpd@sha256:dd993a2108430ec8fdc4942f791ccf9b0c7a6df196907a80e7e8a5f8f1bbf678 -Bbn ${TEST_REGISTRY_USERNAME} ${TEST_REGISTRY_PASSWORD}> ~/auth/htpasswd

	if [ "$$(docker inspect -f '{{.State.Running}}' "registry" 2>/dev/null || true)" ]; then docker stop registry && docker rm registry; fi
	docker pull ${LOCAL_REGISTRY_IMAGE}
	docker run -d \
		-p 5000:5000 \
		--restart=always \
		--name registry \
		-v ${HOME}/auth:/auth \
		-e "REGISTRY_AUTH=htpasswd" \
		-e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
		-e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
		-e REGISTRY_STORAGE_DELETE_ENABLED=true \
		${LOCAL_REGISTRY_IMAGE}
	sleep 5
	docker login -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} ${TEST_REGISTRY}
	${GITHUB_WORKSPACE}/bin/oras login \
		-u ${TEST_REGISTRY_USERNAME} \
		-p ${TEST_REGISTRY_PASSWORD} \
		${TEST_REGISTRY}

e2e-create-all-image:
	rm -rf .staging
	mkdir .staging
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "all-in-one image"]' > .staging/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/all:v0 .staging
	docker push ${TEST_REGISTRY}/all:v0

e2e-bootstrap: e2e-dependencies e2e-create-local-registry
	printf 'kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\ncontainerdConfigPatches:\n- |-\n  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]\n    endpoint = ["http://registry:5000"]' > kind_config.yaml

	# Check for existing kind cluster
	if [ $$(${GITHUB_WORKSPACE}/bin/kind get clusters) ]; then ${GITHUB_WORKSPACE}/bin/kind delete cluster; fi
	# Create a new kind cluster
	TERM=dumb ${GITHUB_WORKSPACE}/bin/kind create cluster --image $(KIND_NODE_VERSION) --wait 5m --config=kind_config.yaml
	if [ "$$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "registry")" = 'null' ]; then docker network connect "kind" "registry"; fi
	rm kind_config.yaml

e2e-helm-install:
	rm -rf .staging/helm
	mkdir -p .staging/helm
	curl https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz --output .staging/helm/helmbin.tar.gz
	cd .staging/helm && tar -xvf helmbin.tar.gz
	./.staging/helm/linux-amd64/helm version --client

e2e-helmfile-install:
	rm -rf .staging/helmfilebin
	mkdir -p .staging/helmfilebin
	curl -L https://github.com/helmfile/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION}_linux_amd64.tar.gz --output .staging/helmfilebin/helmfilebin.tar.gz
	cd .staging/helmfilebin && tar -xvf helmfilebin.tar.gz
    
e2e-docker-credential-store-setup:
	rm -rf .staging/pass
	mkdir -p .staging/pass
	cd .staging/pass && git clone https://github.com/docker/docker-credential-helpers.git
	cd .staging/pass/docker-credential-helpers && make pass
	cp .staging/pass/docker-credential-helpers/bin/build/docker-credential-pass /usr/local/bin/
	sed -i '0,/{/s/{/{\n\t"credsStore": "pass",/' ~/.docker/config.json

	gpg --batch --passphrase '' --quick-gen-key ratify default default
	pass init $$(gpg --list-keys ratify | sed -n 2p | tr -d " \t\n\r")

e2e-notation-setup:
	rm -rf .staging/notation
	mkdir -p .staging/notation
	curl -L https://github.com/notaryproject/notation/releases/download/v${NOTATION_VERSION}/notation_${NOTATION_VERSION}_linux_amd64.tar.gz --output ${GITHUB_WORKSPACE}/.staging/notation/notation.tar.gz
	tar -zxvf ${GITHUB_WORKSPACE}/.staging/notation/notation.tar.gz -C ${GITHUB_WORKSPACE}/.staging/notation

	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "notation signed image"]' > .staging/notation/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/notation:signed .staging/notation
	docker push ${TEST_REGISTRY}/notation:signed

	docker pull ${LOCAL_UNSIGNED_IMAGE}
	docker image tag ${LOCAL_UNSIGNED_IMAGE} ${TEST_REGISTRY}/notation:unsigned
	docker push ${TEST_REGISTRY}/notation:unsigned

	rm -rf ~/.config/notation
	.staging/notation/notation cert generate-test --default "ratify-bats-test"

	.staging/notation/notation sign -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} `docker image inspect ${TEST_REGISTRY}/notation:signed | jq -r .[0].RepoDigests[0]`
	.staging/notation/notation sign -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} `docker image inspect ${TEST_REGISTRY}/all:v0 | jq -r .[0].RepoDigests[0]`

e2e-notation-leaf-cert-setup:
	mkdir -p .staging/notation/leaf-test
	mkdir -p ~/.config/notation/truststore/x509/ca/leaf-test
	./scripts/generate-cert-chain.sh .staging/notation/leaf-test
	cp .staging/notation/leaf-test/leaf.crt ~/.config/notation/truststore/x509/ca/leaf-test/leaf.crt
	cp .staging/notation/leaf-test/ca.crt ~/.config/notation/truststore/x509/ca/leaf-test/root.crt
	cat .staging/notation/leaf-test/ca.crt >> .staging/notation/leaf-test/leaf.crt

	jq '.keys += [{"name":"leaf-test","keyPath":".staging/notation/leaf-test/leaf.key","certPath":".staging/notation/leaf-test/leaf.crt"}]' ~/.config/notation/signingkeys.json > tmp && mv tmp ~/.config/notation/signingkeys.json

	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "notation leaf signed image"]' > .staging/notation/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/notation:leafSigned .staging/notation
	docker push ${TEST_REGISTRY}/notation:leafSigned
	.staging/notation/notation sign -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} --key "leaf-test" `docker image inspect ${TEST_REGISTRY}/notation:leafSigned | jq -r .[0].RepoDigests[0]`

e2e-cosign-setup:
	rm -rf .staging/cosign
	mkdir -p .staging/cosign
	curl -sSLO https://github.com/sigstore/cosign/releases/download/v${COSIGN_VERSION}/cosign-linux-amd64
	mv cosign-linux-amd64 .staging/cosign
	chmod +x .staging/cosign/cosign-linux-amd64

	# image signed with a key
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "cosign signed image"]' > .staging/cosign/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/cosign:signed-key .staging/cosign
	docker push ${TEST_REGISTRY}/cosign:signed-key

	docker pull ${LOCAL_UNSIGNED_IMAGE}
	docker image tag ${LOCAL_UNSIGNED_IMAGE} ${TEST_REGISTRY}/cosign:unsigned
	docker push ${TEST_REGISTRY}/cosign:unsigned

	export COSIGN_PASSWORD="test" && \
	cd .staging/cosign && \
	./cosign-linux-amd64 login ${TEST_REGISTRY} -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} && \
	./cosign-linux-amd64 generate-key-pair && \
	./cosign-linux-amd64 sign --key cosign.key `docker image inspect ${TEST_REGISTRY}/cosign:signed-key | jq -r .[0].RepoDigests[0]` && \
	./cosign-linux-amd64 sign --key cosign.key `docker image inspect ${TEST_REGISTRY}/all:v0 | jq -r .[0].RepoDigests[0]`

e2e-licensechecker-setup:
	rm -rf .staging/licensechecker
	mkdir -p .staging/licensechecker

	# Install Syft
	curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b .staging/licensechecker ${SYFT_VERSION}

	# Build/Push Image
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "licensechecker image"]' > .staging/licensechecker/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/licensechecker:v0 .staging/licensechecker
	docker push ${TEST_REGISTRY}/licensechecker:v0

	# Create/Attach SPDX
	.staging/licensechecker/syft -o spdx --file .staging/licensechecker/sbom.spdx ${TEST_REGISTRY}/licensechecker:v0
	${GITHUB_WORKSPACE}/bin/oras attach ${TEST_REGISTRY}/licensechecker:v0 \
  		--artifact-type application/vnd.ratify.spdx.v0 \
  		.staging/licensechecker/sbom.spdx:application/text
	${GITHUB_WORKSPACE}/bin/oras attach ${TEST_REGISTRY}/all:v0 \
  		--artifact-type application/vnd.ratify.spdx.v0 \
  		.staging/licensechecker/sbom.spdx:application/text

e2e-sbom-setup:
	rm -rf .staging/sbom
	mkdir -p .staging/sbom

	# Install sbom-tool
	curl -Lo .staging/sbom/sbom-tool https://github.com/microsoft/sbom-tool/releases/download/${SBOM_TOOL_VERSION}/sbom-tool-linux-x64 && chmod +x .staging/sbom/sbom-tool
	
	# Install syft
	curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b .staging/sbom ${SYFT_VERSION}

	# Build/Push Images
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "sbom image"]' > .staging/sbom/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/sbom:v0 .staging/sbom
	docker push ${TEST_REGISTRY}/sbom:v0
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "sbom image unsigned"]' > .staging/sbom/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/sbom:unsigned .staging/sbom
	docker push ${TEST_REGISTRY}/sbom:unsigned

	# Generate/Attach sbom
	.staging/sbom/sbom-tool generate -b .staging/sbom -bc . -pn ratify -m .staging/sbom -pv 1.0 -ps acme -nsu ratify -nsb http://registry:5000 -D true
	${GITHUB_WORKSPACE}/bin/oras attach \
		--artifact-type org.example.sbom.v0 \
		 ${TEST_REGISTRY}/sbom:v0 \
		.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json
	${GITHUB_WORKSPACE}/bin/oras attach \
		--artifact-type org.example.sbom.v0 \
		 ${TEST_REGISTRY}/sbom:unsigned \
		.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json
	${GITHUB_WORKSPACE}/bin/oras attach \
		--artifact-type org.example.sbom.v0 \
		 ${TEST_REGISTRY}/all:v0 \
		.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json

	# Generate/Attach sbom with license info
	 .staging/sbom/syft -o spdx-json --file .staging/sbom/sbom.spdx.json ${TEST_REGISTRY}/sbom:v0
	 ${GITHUB_WORKSPACE}/bin/oras attach \
	 	--artifact-type org.example.sbom.v1 \
	 	${TEST_REGISTRY}/all:v0 \
	 	.staging/sbom/sbom.spdx.json:application/spdx+json

	# Push Signature to sbom
	.staging/notation/notation sign -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} ${TEST_REGISTRY}/sbom@`oras discover -o json --artifact-type org.example.sbom.v0 ${TEST_REGISTRY}/sbom:v0 | jq -r ".manifests[0].digest"`
	.staging/notation/notation sign -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} ${TEST_REGISTRY}/all@`oras discover -o json --artifact-type org.example.sbom.v0 ${TEST_REGISTRY}/all:v0 | jq -r ".manifests[0].digest"` 

e2e-schemavalidator-setup:
	rm -rf .staging/schemavalidator
	mkdir -p .staging/schemavalidator

	# Install Trivy
	curl -L https://github.com/aquasecurity/trivy/releases/download/v0.35.0/trivy_0.35.0_Linux-64bit.tar.gz --output .staging/schemavalidator/trivy.tar.gz
	tar -zxf .staging/schemavalidator/trivy.tar.gz -C .staging/schemavalidator

	# Build/Push Images
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "schemavalidator image"]' > .staging/schemavalidator/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/schemavalidator:v0 .staging/schemavalidator
	docker push ${TEST_REGISTRY}/schemavalidator:v0

	# Create/Attach Scan Results
	.staging/schemavalidator/trivy image --format sarif --output .staging/schemavalidator/trivy-scan.sarif ${TEST_REGISTRY}/schemavalidator:v0
	${GITHUB_WORKSPACE}/bin/oras attach \
		--artifact-type vnd.aquasecurity.trivy.report.sarif.v1 \
		${TEST_REGISTRY}/schemavalidator:v0 \
		.staging/schemavalidator/trivy-scan.sarif:application/sarif+json
	${GITHUB_WORKSPACE}/bin/oras attach \
		--artifact-type vnd.aquasecurity.trivy.report.sarif.v1 \
		${TEST_REGISTRY}/all:v0 \
		.staging/schemavalidator/trivy-scan.sarif:application/sarif+json

e2e-inlinecert-setup:
	rm -rf .staging/inlinecert
	mkdir -p .staging/inlinecert

	# build and sign an image with an alternate certificate
	printf 'FROM ${ALPINE_IMAGE}\nCMD ["echo", "alternate notation signed image"]' > .staging/inlinecert/Dockerfile
	docker build --no-cache -t ${TEST_REGISTRY}/notation:signed-alternate .staging/inlinecert
	docker push ${TEST_REGISTRY}/notation:signed-alternate

	.staging/notation/notation cert generate-test "alternate-cert"
	.staging/notation/notation sign -u ${TEST_REGISTRY_USERNAME} -p ${TEST_REGISTRY_PASSWORD} --key "alternate-cert" `docker image inspect ${TEST_REGISTRY}/notation:signed-alternate | jq -r .[0].RepoDigests[0]`

e2e-azure-setup: e2e-create-all-image e2e-notation-setup e2e-notation-leaf-cert-setup e2e-cosign-setup e2e-licensechecker-setup e2e-sbom-setup e2e-schemavalidator-setup

e2e-deploy-gatekeeper: e2e-helm-install
	./.staging/helm/linux-amd64/helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
	if [ ${GATEKEEPER_VERSION} = "3.12.0" ] || [ ${GATEKEEPER_VERSION} = "3.13.0" ]; then ./.staging/helm/linux-amd64/helm install gatekeeper/gatekeeper --version ${GATEKEEPER_VERSION} --name-template=gatekeeper --namespace ${GATEKEEPER_NAMESPACE} --create-namespace --set enableExternalData=true --set validatingWebhookTimeoutSeconds=5 --set mutatingWebhookTimeoutSeconds=2 --set auditInterval=0; fi
	if [ ${GATEKEEPER_VERSION} = "3.13.0" ]; then kubectl -n ${GATEKEEPER_NAMESPACE} patch deployment gatekeeper-controller-manager --type=json -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--external-data-provider-response-cache-ttl=1s"}]' && sleep 60; fi
	# Gatekeeper versions >= 3.14.0 need a special helm value to override the default external data response cache ttl to 10s
	if [ ${GATEKEEPER_VERSION} != "3.12.0" ] && [ ${GATEKEEPER_VERSION} != "3.13.0" ]; then ./.staging/helm/linux-amd64/helm install gatekeeper/gatekeeper --version ${GATEKEEPER_VERSION} --name-template=gatekeeper --namespace ${GATEKEEPER_NAMESPACE} --create-namespace --set enableExternalData=true --set validatingWebhookTimeoutSeconds=5 --set mutatingWebhookTimeoutSeconds=2 --set auditInterval=0 --set externaldataProviderResponseCacheTTL=1s; fi

e2e-build-crd-image:
	docker build --progress=plain --no-cache --build-arg KUBE_VERSION=${KUBERNETES_VERSION} --build-arg TARGETOS="linux" --build-arg TARGETARCH="amd64" -f crd.Dockerfile -t localbuildcrd:test ./charts/ratify/crds
	kind load docker-image --name kind localbuildcrd:test

e2e-deploy-base-ratify: e2e-notation-setup e2e-notation-leaf-cert-setup e2e-inlinecert-setup e2e-build-crd-image
	docker build --progress=plain --no-cache \
	-f ./httpserver/Dockerfile \
	-t baselocalbuild:test .
	kind load docker-image --name kind baselocalbuild:test

	printf "{\n\t\"auths\": {\n\t\t\"registry:5000\": {\n\t\t\t\"auth\": \"`echo "${TEST_REGISTRY_USERNAME}:${TEST_REGISTRY_PASSWORD}" | tr -d '\n' | base64 -i -w 0`\"\n\t\t}\n\t}\n}" > mount_config.json

	./.staging/helm/linux-amd64/helm install ${RATIFY_NAME} \
    ./charts/ratify --atomic --namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
	--set image.repository=baselocalbuild \
	--set image.crdRepository=localbuildcrd \
	--set image.tag=test \
	--set gatekeeper.version=${GATEKEEPER_VERSION} \
	--set featureFlags.RATIFY_CERT_ROTATION=${CERT_ROTATION_ENABLED} \
	--set-file provider.tls.crt=${CERT_DIR}/server.crt \
	--set-file provider.tls.key=${CERT_DIR}/server.key \
	--set-file provider.tls.caCert=${CERT_DIR}/ca.crt \
    --set-file provider.tls.caKey=${CERT_DIR}/ca.key \
	--set provider.tls.cabundle="$(shell cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
	--set notationCerts[0]="$$(cat ~/.config/notation/localkeys/ratify-bats-test.crt)" \
	--set oras.useHttp=true \
	--set cosign.enabled=false \
	--set-file dockerConfig="mount_config.json" \
	--set logger.level=debug

	rm mount_config.json

e2e-deploy-ratify: e2e-notation-setup e2e-notation-leaf-cert-setup e2e-cosign-setup e2e-cosign-setup e2e-licensechecker-setup e2e-sbom-setup e2e-schemavalidator-setup e2e-inlinecert-setup e2e-build-crd-image e2e-build-local-ratify-image e2e-helm-deploy-ratify

e2e-build-local-ratify-image:
	docker build --progress=plain --no-cache \
	--build-arg build_cosign=true \
	--build-arg build_sbom=true \
	--build-arg build_licensechecker=true \
	--build-arg build_schemavalidator=true \
	-f ./httpserver/Dockerfile \
	-t localbuild:test .
	kind load docker-image --name kind localbuild:test

e2e-helmfile-deploy-released-ratify:
	./.staging/helmfilebin/helmfile sync -f git::https://github.com/deislabs/ratify.git@helmfile.yaml

e2e-helm-deploy-ratify:
	printf "{\n\t\"auths\": {\n\t\t\"registry:5000\": {\n\t\t\t\"auth\": \"`echo "${TEST_REGISTRY_USERNAME}:${TEST_REGISTRY_PASSWORD}" | tr -d '\n' | base64 -i -w 0`\"\n\t\t}\n\t}\n}" > mount_config.json

	./.staging/helm/linux-amd64/helm install ${RATIFY_NAME} \
    ./charts/ratify --atomic --namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
	--set image.repository=localbuild \
	--set image.crdRepository=localbuildcrd \
	--set image.tag=test \
	--set gatekeeper.version=${GATEKEEPER_VERSION} \
	--set featureFlags.RATIFY_CERT_ROTATION=${CERT_ROTATION_ENABLED} \
	--set-file provider.tls.crt=${CERT_DIR}/server.crt \
	--set-file provider.tls.key=${CERT_DIR}/server.key \
	--set-file provider.tls.caCert=${CERT_DIR}/ca.crt \
    --set-file provider.tls.caKey=${CERT_DIR}/ca.key \
	--set provider.tls.cabundle="$(shell cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
	--set notationCerts[0]="$$(cat ~/.config/notation/localkeys/ratify-bats-test.crt)" \
	--set cosign.key="$$(cat .staging/cosign/cosign.pub)" \
	--set oras.useHttp=true \
	--set-file dockerConfig="mount_config.json" \
	--set logger.level=debug

	rm mount_config.json

e2e-helm-deploy-ratify-without-tls-certs:
	printf "{\n\t\"auths\": {\n\t\t\"registry:5000\": {\n\t\t\t\"auth\": \"`echo "${TEST_REGISTRY_USERNAME}:${TEST_REGISTRY_PASSWORD}" | tr -d '\n' | base64 -i -w 0`\"\n\t\t}\n\t}\n}" > mount_config.json

	./.staging/helm/linux-amd64/helm install ${RATIFY_NAME} \
    ./charts/ratify --atomic --namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
	--set image.repository=localbuild \
	--set image.crdRepository=localbuildcrd \
	--set image.tag=test \
	--set gatekeeper.version=${GATEKEEPER_VERSION} \
	--set featureFlags.RATIFY_CERT_ROTATION=${CERT_ROTATION_ENABLED} \
	--set notaryCert="$$(cat ~/.config/notation/localkeys/ratify-bats-test.crt)" \
	--set cosign.key="$$(cat .staging/cosign/cosign.pub)" \
	--set oras.useHttp=true \
	--set-file dockerConfig="mount_config.json" \
	--set logger.level=debug

	rm mount_config.json

e2e-helm-deploy-dapr:
	helm repo add dapr https://dapr.github.io/helm-charts/
	helm repo update
	helm upgrade --install --version ${DAPR_VERSION} dapr dapr/dapr --namespace dapr-system --create-namespace --wait

e2e-helm-deploy-redis: e2e-helm-deploy-dapr
	# overwrite pod security labels to baseline
	kubectl label --overwrite ns ${GATEKEEPER_NAMESPACE} pod-security.kubernetes.io/enforce=baseline
	kubectl label --overwrite ns ${GATEKEEPER_NAMESPACE} pod-security.kubernetes.io/warn=baseline
	
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	helm upgrade --install redis bitnami/redis --namespace ${GATEKEEPER_NAMESPACE} --set image.tag=${REDIS_IMAGE_TAG} --wait --set replica.replicaCount=1 --set tls.enabled=true --set tls.autoGenerated=true --set tls.authClients=false
	SIGN_KEY=$(shell openssl rand 16 | hexdump -v -e '/1 "%02x"' | base64) ${GITHUB_WORKSPACE}/bin/yq -i '.data.signingKey = strenv(SIGN_KEY)' test/testdata/dapr/dapr-redis-secret.yaml
	kubectl apply -f test/testdata/dapr/dapr-redis-secret.yaml -n ${GATEKEEPER_NAMESPACE}
	kubectl apply -f test/testdata/dapr/dapr-redis.yaml -n ${GATEKEEPER_NAMESPACE}
 
e2e-helm-deploy-ratify-replica: e2e-helm-deploy-redis e2e-notation-setup e2e-build-crd-image e2e-build-local-ratify-image
	printf "{\n\t\"auths\": {\n\t\t\"registry:5000\": {\n\t\t\t\"auth\": \"`echo "${TEST_REGISTRY_USERNAME}:${TEST_REGISTRY_PASSWORD}" | tr -d '\n' | base64 -i -w 0`\"\n\t\t}\n\t}\n}" > mount_config.json

	./.staging/helm/linux-amd64/helm install ${RATIFY_NAME} \
    ./charts/ratify --atomic --namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
	--set image.repository=localbuild \
	--set image.crdRepository=localbuildcrd \
	--set image.tag=test \
	--set gatekeeper.version=${GATEKEEPER_VERSION} \
	--set featureFlags.RATIFY_CERT_ROTATION=${CERT_ROTATION_ENABLED} \
	--set-file provider.tls.crt=${CERT_DIR}/server.crt \
	--set-file provider.tls.key=${CERT_DIR}/server.key \
	--set-file provider.tls.caCert=${CERT_DIR}/ca.crt \
    --set-file provider.tls.caKey=${CERT_DIR}/ca.key \
	--set provider.tls.cabundle="$(shell cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
	--set notationCerts[0]="$$(cat ~/.config/notation/localkeys/ratify-bats-test.crt)" \
	--set oras.useHttp=true \
	--set cosign.enabled=false \
	--set-file dockerConfig="mount_config.json" \
	--set logger.level=debug \
	--set replicaCount=2 \
	--set provider.cache.type="dapr" \
	--set provider.cache.name="dapr-redis" \
	--set featureFlags.RATIFY_EXPERIMENTAL_HIGH_AVAILABILITY=true \
	--set resources.requests.memory="64Mi" \
	--set resources.requests.cpu="200m"

	rm mount_config.json

e2e-aks:
	./scripts/azure-ci-test.sh ${KUBERNETES_VERSION} ${GATEKEEPER_VERSION} ${TENANT_ID} ${GATEKEEPER_NAMESPACE} ${CERT_DIR}

e2e-cleanup:
	./scripts/azure-ci-test-cleanup.sh ${AZURE_SUBSCRIPTION_ID}
##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen conversion-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations. Also generate conversions between structs of different API versions.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	$(CONVERSION_GEN) \
        --input-dirs "./api/v1beta1,./api/v1alpha1" \
        --go-header-file "./hack/boilerplate.go.txt" \
        --output-file-base "zz_generated.conversion"

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install-crds
install-crds: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall-crds
uninstall-crds: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
CONVERSION_GEN ?= $(LOCALBIN)/conversion-gen

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.2
CONVERSION_TOOLS_VERSION ?= v0.26.1

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: conversion-gen
conversion-gen: $(CONVERSION_GEN) ## Download conversion-gen locally if necessary.
$(CONVERSION_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/conversion-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/conversion-gen@$(CONVERSION_TOOLS_VERSION)
