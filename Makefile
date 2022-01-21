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