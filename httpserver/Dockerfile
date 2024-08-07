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

FROM --platform=$BUILDPLATFORM golang:1.22@sha256:86a3c48a61915a8c62c0e1d7594730399caa3feb73655dfe96c7bc17710e96cf as builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG LDFLAGS
ARG GOPROXY
ARG build_sbom
ARG build_licensechecker
ARG build_schemavalidator
ARG build_vulnerabilityreport

ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT} \
    GOPROXY=${GOPROXY}

WORKDIR /app

COPY . .

RUN go build -ldflags "${LDFLAGS}" -o /app/out/ /app/cmd/ratify
RUN mkdir /app/out/plugins
RUN if [ "$build_sbom" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/sbom; fi
RUN if [ "$build_licensechecker" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/licensechecker; fi
RUN if [ "$build_schemavalidator" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/schemavalidator; fi
RUN if [ "$build_vulnerabilityreport" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/vulnerabilityreport; fi

FROM gcr.io/distroless/static:nonroot@sha256:8dd8d3ca2cf283383304fd45a5c9c74d5f2cd9da8d3b077d720e264880077c65
LABEL org.opencontainers.image.source https://github.com/ratify-project/ratify

ARG RATIFY_FOLDER=$HOME/.ratify/

WORKDIR /

COPY --from=builder /app/out/ratify /app/
COPY --from=builder --chown=65532:65532 /app/out/plugins ${RATIFY_FOLDER}/plugins
COPY --from=builder /app/config/config.json ${RATIFY_FOLDER}

ENV RATIFY_CONFIG=${RATIFY_FOLDER}

EXPOSE 6001
EXPOSE 8888

USER 65532:65532

ENTRYPOINT ["/app/ratify", "serve", "--http", ":6001"]

