BINARY_NAME=ratify
INSTALL_DIR=~/.ratify
all: build test

.PHONY: build
build: build-cli build-plugins

.PHONY: build-cli 
build-cli:
	go build -o ./bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

.PHONY: build-plugins
build-plugins: 
	go build -o ./bin/plugins/ ./plugins/verifier/cosign
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
