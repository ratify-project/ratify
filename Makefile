BINARY_NAME=ratify
INSTALL_DIR=~/.ratify
all: build test

build: build-cli build-plugins

build-cli:
	go build -o ./bin/${BINARY_NAME} ./cmd/ratify


build-plugins: 
	go build -o ./bin/plugins/ ./plugins/verifier/sbom

install: 
	mkdir -p ${INSTALL_DIR}
	cp -r ./bin/* ${INSTALL_DIR}

test:
	go test -v ./...

clean:
	go clean
	rm ./bin/${BINARY_NAME}
