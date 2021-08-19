BINARY_NAME=hora
INSTALL_DIR=~/.hora
all: build test

build: build-cli build-plugins

build-cli:
	go build -o ./bin/${BINARY_NAME} ./cmd/hora


build-plugins: 
	go build -o ./bin/plugins/ ./plugins/referrerstore/ociregistry
	go build -o ./bin/plugins/ ./plugins/verifier/nv2verifier
	go build -o ./bin/plugins/ ./plugins/verifier/sbom

install: 
	mkdir -p ${INSTALL_DIR}
	cp -r ./bin/* ${INSTALL_DIR}

test:
	go test -v ./cmd/hora
 
clean:
	go clean
	rm ./bin/${BINARY_NAME}
