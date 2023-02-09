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
KUBERNETES_VERSION ?= 1.25.4
GATEKEEPER_VERSION ?= 3.11.0
COSIGN_VERSION ?= 1.13.1

HELM_VERSION ?= 3.9.2
BATS_TESTS_FILE ?= test/bats/test.bats
BATS_CLI_TESTS_FILE ?= test/bats/cli-test.bats
BATS_VERSION ?= 1.7.0

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.2

GATEKEEPER_NAMESPACE = gatekeeper-system
RATIFY_NAME = ratify

# Local Registry Setup
LOCAL_REGISTRY_IMAGE ?= ghcr.io/oras-project/registry:v1.0.0-rc.4
LOCAL_UNSIGNED_IMAGE = hello-world:latest
LOCAL_TEST_REGISTRY = localhost:5000

all: build test

.PHONY: build
build: build-cli build-plugins

.PHONY: build-cli
build-cli: generate fmt vet
	go build --ldflags="$(LDFLAGS)" \
	-o ./bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

.PHONY: build-plugins
build-plugins:
	go build -o ./bin/plugins/ ./plugins/verifier/cosign
	go build -o ./bin/plugins/ ./plugins/verifier/licensechecker
	go build -o ./bin/plugins/ ./plugins/verifier/sample
	go build -o ./bin/plugins/ ./plugins/verifier/sbom
	go build -o ./bin/plugins/ ./plugins/verifier/schemavalidator

.PHONY: install
install:
	mkdir -p ${INSTALL_DIR}
	mkdir -p ${INSTALL_DIR}/ratify-certs/cosign
	mkdir -p ${INSTALL_DIR}/ratify-certs/notary
	cp -r ./bin/* ${INSTALL_DIR}

.PHONY: ratify-config
ratify-config:
	cp ./test/bats/tests/config/* ${INSTALL_DIR}
	cp ./test/bats/tests/certificates/wabbit-networks.io.crt ${INSTALL_DIR}/ratify-certs/notary/wabbit-networks.io.crt
	cp ./test/bats/tests/certificates/cosign.pub ${INSTALL_DIR}/ratify-certs/cosign/cosign.pub

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

.PHONY: deploy-gatekeeper
deploy-gatekeeper:
	helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
	helm install gatekeeper/gatekeeper  \
		--version ${GATEKEEPER_VERSION} \
		--name-template=gatekeeper \
		--namespace ${GATEKEEPER_NAMESPACE} --create-namespace

.PHONY: delete-gatekeeper
delete-gatekeeper:
	helm delete gatekeeper --namespace ${GATEKEEPER_NAMESPACE}

.PHONY: test-e2e
test-e2e:
	bats -t ${BATS_TESTS_FILE}

.PHONY: test-e2e-cli

test-e2e-cli:e2e-dependencies e2e-create-local-registry e2e-notaryv2-setup e2e-cosign-setup e2e-licensechecker-setup e2e-sbom-setup e2e-schemavalidator-setup
	RATIFY_DIR=${INSTALL_DIR} LOCAL_TEST_REGISTRY=${LOCAL_TEST_REGISTRY} ${GITHUB_WORKSPACE}/bin/bats -t ${BATS_CLI_TESTS_FILE}

.PHONY: generate-certs
generate-certs:
	./scripts/generate-tls-certs.sh ${CERT_DIR} ${GATEKEEPER_NAMESPACE}

install-bats:
	# Download and install bats
	curl -sSLO https://github.com/bats-core/bats-core/archive/v${BATS_VERSION}.tar.gz && tar -zxvf v${BATS_VERSION}.tar.gz && bash bats-core-${BATS_VERSION}/install.sh ${GITHUB_WORKSPACE}

e2e-dependencies:
	# Download and install kind
	curl -L https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-linux-amd64 --output ${GITHUB_WORKSPACE}/bin/kind && chmod +x ${GITHUB_WORKSPACE}/bin/kind
	# Download and install kubectl
	curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl --output ${GITHUB_WORKSPACE}/bin/kubectl && chmod +x ${GITHUB_WORKSPACE}/bin/kubectl
	# Download and install bats
	curl -sSLO https://github.com/bats-core/bats-core/archive/v${BATS_VERSION}.tar.gz && tar -zxvf v${BATS_VERSION}.tar.gz && bash bats-core-${BATS_VERSION}/install.sh ${GITHUB_WORKSPACE}
	# Download and install jq
	curl -L https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 --output ${GITHUB_WORKSPACE}/bin/jq && chmod +x ${GITHUB_WORKSPACE}/bin/jq
	# Install ORAS
	curl -LO https://github.com/oras-project/oras/releases/download/v0.16.0/oras_0.16.0_linux_amd64.tar.gz
	mkdir -p oras-install/
	tar -zxf oras_0.16.0_*.tar.gz -C oras-install/
	mv oras-install/oras ${GITHUB_WORKSPACE}/bin
	rm -rf oras_0.16.0_*.tar.gz oras-install/

KIND_NODE_VERSION := kindest/node:v$(KUBERNETES_VERSION)

e2e-create-local-registry:
	if [ "$$(docker inspect -f '{{.State.Running}}' "registry" 2>/dev/null || true)" ]; then docker stop registry && docker rm registry; fi
	docker pull ${LOCAL_REGISTRY_IMAGE}
	docker run -d -p 5000:5000 --restart=always --name registry ${LOCAL_REGISTRY_IMAGE}

	rm -rf .staging
	mkdir .staging
	echo 'FROM alpine\nCMD ["echo", "all-in-one image"]' > .staging/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/all:v0 .staging
	docker push ${LOCAL_TEST_REGISTRY}/all:v0

e2e-bootstrap: e2e-dependencies e2e-create-local-registry
	echo 'kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\ncontainerdConfigPatches:\n- |-\n  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]\n    endpoint = ["http://registry:5000"]' > kind_config.yaml

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

e2e-notaryv2-setup:
	rm -rf .staging/notaryv2
	mkdir -p .staging/notaryv2
	cd .staging/notaryv2 && git clone https://github.com/notaryproject/notation
	cd .staging/notaryv2/notation && make build

	echo 'FROM alpine\nCMD ["echo", "notaryv2 signed image"]' > .staging/notaryv2/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/notation:signed .staging/notaryv2
	docker push ${LOCAL_TEST_REGISTRY}/notation:signed

	docker pull ${LOCAL_UNSIGNED_IMAGE}
	docker image tag ${LOCAL_UNSIGNED_IMAGE} ${LOCAL_TEST_REGISTRY}/notation:unsigned
	docker push ${LOCAL_TEST_REGISTRY}/notation:unsigned

	rm -rf ~/.config/notation
	.staging/notaryv2/notation/bin/notation cert generate-test --default "ratify-bats-test"
	.staging/notaryv2/notation/bin/notation sign `docker image inspect ${LOCAL_TEST_REGISTRY}/notation:signed | jq -r .[0].RepoDigests[0]`
	.staging/notaryv2/notation/bin/notation sign `docker image inspect ${LOCAL_TEST_REGISTRY}/all:v0 | jq -r .[0].RepoDigests[0]`

e2e-cosign-setup:
	rm -rf .staging/cosign
	mkdir -p .staging/cosign
	curl -sSLO https://github.com/sigstore/cosign/releases/download/v${COSIGN_VERSION}/cosign-linux-amd64
	mv cosign-linux-amd64 .staging/cosign
	chmod +x .staging/cosign/cosign-linux-amd64

	# image signed with a key
	echo 'FROM alpine\nCMD ["echo", "cosign signed image"]' > .staging/cosign/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/cosign:signed-key .staging/cosign
	docker push ${LOCAL_TEST_REGISTRY}/cosign:signed-key

	# image signed with fulcio root CA
	echo 'FROM alpine\nCMD ["echo", "cosign keyless signed image"]' > .staging/cosign/Dockerfile.keyless
	docker build -t ${LOCAL_TEST_REGISTRY}/cosign:signed-keyless .staging/cosign -f .staging/cosign/Dockerfile.keyless
	docker push ${LOCAL_TEST_REGISTRY}/cosign:signed-keyless

	docker pull ${LOCAL_UNSIGNED_IMAGE}
	docker image tag ${LOCAL_UNSIGNED_IMAGE} ${LOCAL_TEST_REGISTRY}/cosign:unsigned
	docker push ${LOCAL_TEST_REGISTRY}/cosign:unsigned

	export COSIGN_PASSWORD="test" && \
	cd .staging/cosign && \
	./cosign-linux-amd64 generate-key-pair && \
	./cosign-linux-amd64 sign --key cosign.key `docker image inspect ${LOCAL_TEST_REGISTRY}/cosign:signed-key | jq -r .[0].RepoDigests[0]` && \
	COSIGN_EXPERIMENTAL=1 ./cosign-linux-amd64 sign `docker image inspect ${LOCAL_TEST_REGISTRY}/cosign:signed-keyless | jq -r .[0].RepoDigests[0]` && \
	COSIGN_EXPERIMENTAL=1 ./cosign-linux-amd64 sign `docker image inspect ${LOCAL_TEST_REGISTRY}/all:v0 | jq -r .[0].RepoDigests[0]`

e2e-licensechecker-setup:
	rm -rf .staging/licensechecker
	mkdir -p .staging/licensechecker

	# Install Syft
	curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b .staging/licensechecker

	# Build/Push Image
	echo 'FROM alpine@sha256:93d5a28ff72d288d69b5997b8ba47396d2cbb62a72b5d87cd3351094b5d578a0\nCMD ["echo", "licensechecker image"]' > .staging/licensechecker/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/licensechecker:v0 .staging/licensechecker
	docker push ${LOCAL_TEST_REGISTRY}/licensechecker:v0

	# Create/Attach SPDX
	.staging/licensechecker/syft -o spdx --file .staging/licensechecker/sbom.spdx ${LOCAL_TEST_REGISTRY}/licensechecker:v0
	oras attach ${LOCAL_TEST_REGISTRY}/licensechecker:v0 \
  		--artifact-type application/vnd.ratify.spdx.v0 \
  		--plain-http \
  		.staging/licensechecker/sbom.spdx:application/text
	oras attach ${LOCAL_TEST_REGISTRY}/all:v0 \
  		--artifact-type application/vnd.ratify.spdx.v0 \
  		--plain-http \
  		.staging/licensechecker/sbom.spdx:application/text

e2e-sbom-setup:
	rm -rf .staging/sbom
	mkdir -p .staging/sbom

	# Install sbom-tool
	curl -Lo .staging/sbom/sbom-tool https://github.com/microsoft/sbom-tool/releases/latest/download/sbom-tool-linux-x64 && chmod +x .staging/sbom/sbom-tool
	cd .staging/sbom && git clone https://github.com/deislabs/ratify

	# Build/Push Images
	echo 'FROM alpine\nCMD ["echo", "sbom image"]' > .staging/sbom/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/sbom:v0 .staging/sbom
	docker push ${LOCAL_TEST_REGISTRY}/sbom:v0
	echo 'FROM alpine\nCMD ["echo", "sbom image unsigned"]' > .staging/sbom/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/sbom:unsigned .staging/sbom
	docker push ${LOCAL_TEST_REGISTRY}/sbom:unsigned

	# Generate/Attach sbom
	.staging/sbom/sbom-tool generate -b .staging/sbom -bc .staging/sbom/ratify -pn ratify -m .staging/sbom -pv 1.0 -ps acme -nsu ratify -nsb http://registry:5000 -D true
	oras attach \
		--artifact-type org.example.sbom.v0 \
		--plain-http \
		 ${LOCAL_TEST_REGISTRY}/sbom:v0 \
		.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json
	oras attach \
		--artifact-type org.example.sbom.v0 \
		--plain-http \
		 ${LOCAL_TEST_REGISTRY}/sbom:unsigned \
		.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json
	oras attach \
		--artifact-type org.example.sbom.v0 \
		--plain-http \
		 ${LOCAL_TEST_REGISTRY}/all:v0 \
		.staging/sbom/_manifest/spdx_2.2/manifest.spdx.json:application/spdx+json

	# Push Signature to sbom
	.staging/notaryv2/notation/bin/notation sign  ${LOCAL_TEST_REGISTRY}/sbom@`oras discover -o json --artifact-type org.example.sbom.v0 ${LOCAL_TEST_REGISTRY}/sbom:v0 | jq -r ".manifests[0].digest"`
	.staging/notaryv2/notation/bin/notation sign  ${LOCAL_TEST_REGISTRY}/all@`oras discover -o json --artifact-type org.example.sbom.v0 ${LOCAL_TEST_REGISTRY}/all:v0 | jq -r ".manifests[0].digest"`

e2e-schemavalidator-setup:
	rm -rf .staging/schemavalidator
	mkdir -p .staging/schemavalidator

	# Install Trivy
	curl -L https://github.com/aquasecurity/trivy/releases/download/v0.35.0/trivy_0.35.0_Linux-64bit.tar.gz --output .staging/schemavalidator/trivy.tar.gz
	tar -zxf .staging/schemavalidator/trivy.tar.gz -C .staging/schemavalidator

	# Build/Push Images
	echo 'FROM alpine\nCMD ["echo", "schemavalidator image"]' > .staging/schemavalidator/Dockerfile
	docker build -t ${LOCAL_TEST_REGISTRY}/schemavalidator:v0 .staging/schemavalidator
	docker push ${LOCAL_TEST_REGISTRY}/schemavalidator:v0

	# Create/Attach Scan Results
	.staging/schemavalidator/trivy image --format sarif --output .staging/schemavalidator/trivy-scan.sarif ${LOCAL_TEST_REGISTRY}/schemavalidator:v0
	oras attach \
		--artifact-type vnd.aquasecurity.trivy.report.sarif.v1 \
		${LOCAL_TEST_REGISTRY}/schemavalidator:v0 \
		.staging/schemavalidator/trivy-scan.sarif:application/sarif+json
	oras attach \
		--artifact-type vnd.aquasecurity.trivy.report.sarif.v1 \
		${LOCAL_TEST_REGISTRY}/all:v0 \
		.staging/schemavalidator/trivy-scan.sarif:application/sarif+json

e2e-deploy-gatekeeper: e2e-helm-install
	./.staging/helm/linux-amd64/helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts
	./.staging/helm/linux-amd64/helm install gatekeeper/gatekeeper  \
	--version ${GATEKEEPER_VERSION} \
    --name-template=gatekeeper \
    --namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=7 \
    --set auditInterval=0 \

e2e-deploy-ratify: e2e-notaryv2-setup e2e-cosign-setup e2e-licensechecker-setup e2e-sbom-setup e2e-schemavalidator-setup
	docker build --progress=plain --no-cache -f ./httpserver/Dockerfile -t localbuild:test .
	kind load docker-image --name kind localbuild:test

	docker build --progress=plain --no-cache --build-arg KUBE_VERSION=${KUBERNETES_VERSION} --build-arg TARGETOS="linux" --build-arg TARGETARCH="amd64" -f crd.Dockerfile -t localbuildcrd:test ./charts/ratify/crds
	kind load docker-image --name kind localbuildcrd:test

	./.staging/helm/linux-amd64/helm install ${RATIFY_NAME} \
    ./charts/ratify --atomic --namespace ${GATEKEEPER_NAMESPACE} --create-namespace \
	--set image.repository=localbuild \
	--set image.crdRepository=localbuildcrd \
	--set image.tag=test \
	--set gatekeeper.version=${GATEKEEPER_VERSION} \
	--set-file provider.tls.crt=${CERT_DIR}/server.crt \
	--set-file provider.tls.key=${CERT_DIR}/server.key \
	--set provider.tls.cabundle="$(shell cat ${CERT_DIR}/ca.crt | base64 | tr -d '\n')" \
	--set ratifyTestCert="$$(cat ~/.config/notation/localkeys/ratify-bats-test.crt)" \
	--set cosign.enabled=true \
	--set cosign.key="$$(cat .staging/cosign/cosign.pub)" \
	--set oras.useHttp=true

	kubectl delete verifiers.config.ratify.deislabs.io/verifier-cosign
e2e-aks:
	./scripts/azure-ci-test.sh ${KUBERNETES_VERSION} ${GATEKEEPER_VERSION} ${TENANT_ID} ${GATEKEEPER_NAMESPACE} ${CERT_DIR}

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

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

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
