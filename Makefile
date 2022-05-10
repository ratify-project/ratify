BINARY_NAME		= ratify
INSTALL_DIR		= ~/.ratify

GO_PKG			= github.com/deislabs/ratify
GIT_COMMIT_HASH = $(shell git rev-parse HEAD)
GIT_TREE_STATE 	= $(shell test -n "`git status --porcelain`" && echo "modified" || echo "unmodified")
GIT_TAG     	= $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

LDFLAGS = -w
LDFLAGS += -X $(GO_PKG)/internal/version.GitCommitHash=$(GIT_COMMIT_HASH)
LDFLAGS += -X $(GO_PKG)/internal/version.GitTreeState=$(GIT_TREE_STATE)
LDFLAGS += -X $(GO_PKG)/internal/version.GitTag=$(GIT_TAG)

KIND_VERSION ?= 0.11.0
HELM_VERSION ?= 3.7.2
BATS_TESTS_FILE ?= test/bats/test.bats
BATS_VERSION ?= 1.2.1
YQ_VERSION ?= 4.2.0

all: build test

.PHONY: build
build: build-cli build-plugins

.PHONY: build-cli 
build-cli:
	go build --ldflags="$(LDFLAGS)" \
	-o ./bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

.PHONY: build-plugins
build-plugins: 
	go build -o ./bin/plugins/ ./plugins/verifier/cosign
	go build -o ./bin/plugins/ ./plugins/verifier/licensechecker
	go build -o ./bin/plugins/ ./plugins/verifier/sample
	go build -o ./bin/plugins/ ./plugins/verifier/sbom

.PHONY: install
install: 
	mkdir -p ${INSTALL_DIR}
	cp -r ./bin/* ${INSTALL_DIR}

.PHONY: test
test:
	go test -v ./...

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
	kubectl apply -f ./charts/ratify-gatekeeper/templates/constraint.yaml	 

.PHONY: delete-demo-constraints
delete-demo-constraints:
	kubectl delete -f ./charts/ratify-gatekeeper/templates/constraint.yaml

.PHONY: deploy-gatekeeper
deploy-gatekeeper:
	helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts 
	helm install gatekeeper/gatekeeper  \
	    --name-template=gatekeeper \
	    --namespace gatekeeper-system --create-namespace \
	    --set enableExternalData=true \
	    --set controllerManager.dnsPolicy=ClusterFirst,audit.dnsPolicy=ClusterFirst

.PHONY: delete-gatekeeper
delete-gatekeeper:
	helm delete gatekeeper --namespace gatekeeper-system 

.PHONY: test-e2e
test-e2e:
	bats -t ${BATS_TESTS_FILE}

e2e-dependencies:
	# Download and install kind
	curl -L https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-linux-amd64 --output ${GITHUB_WORKSPACE}/bin/kind && chmod +x ${GITHUB_WORKSPACE}/bin/kind
	# Download and install kubectl
	curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl -o ${GITHUB_WORKSPACE}/bin/kubectl && chmod +x ${GITHUB_WORKSPACE}/bin/kubectl
	# Download and install bats
	curl -sSLO https://github.com/bats-core/bats-core/archive/v${BATS_VERSION}.tar.gz && tar -zxvf v${BATS_VERSION}.tar.gz && bash bats-core-${BATS_VERSION}/install.sh ${GITHUB_WORKSPACE}
	# Install yq
	curl -L https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_linux_amd64 -o ${GITHUB_WORKSPACE}/bin/yq && chmod +x ${GITHUB_WORKSPACE}/bin/yq

KIND_NODE_VERSION := kindest/node:v$(KUBERNETES_VERSION)

e2e-bootstrap: e2e-dependencies
	# Check for existing kind cluster
	if [ $$(${GITHUB_WORKSPACE}/bin/kind get clusters) ]; then ${GITHUB_WORKSPACE}/bin/kind delete cluster; fi
	# Create a new kind cluster
	TERM=dumb ${GITHUB_WORKSPACE}/bin/kind create cluster --image $(KIND_NODE_VERSION) --wait 5m

e2e-helm-install:
	rm -rf .staging/helm
	mkdir -p .staging/helm
	curl https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz > .staging/helm/helmbin.tar.gz
	cd .staging/helm && tar -xvf helmbin.tar.gz
	./.staging/helm/linux-amd64/helm version --client

e2e-deploy-gatekeeper: e2e-helm-install
	./.staging/helm/linux-amd64/helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts 
	./.staging/helm/linux-amd64/helm gatekeeper/gatekeeper  \
    --name-template=gatekeeper \
    --namespace gatekeeper-system --create-namespace \
    --set enableExternalData=true \
    --set validatingWebhookTimeoutSeconds=7

e2e-deploy-ratify:
	docker build -f ./httpserver/Dockerfile -t dummy-provider:test	 
	kind load docker-image --name kind dummy-provider:test
	./.staging/helm/linux-amd64/helm repo add ratify https://deislabs.github.io/ratify
	./.staging/helm/linux-amd64/helm install ratify \
    ratify/ratify --atomic	