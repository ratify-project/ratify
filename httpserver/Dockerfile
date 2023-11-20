ARG BUILDERIMAGE="golang:1.20"
ARG BASEIMAGE="gcr.io/distroless/static:nonroot"

FROM --platform=$BUILDPLATFORM $BUILDERIMAGE as builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG LDFLAGS
ARG GOPROXY
ARG build_sbom
ARG build_cosign
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
RUN if [ "$build_cosign" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/cosign; fi
RUN if [ "$build_licensechecker" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/licensechecker; fi
RUN if [ "$build_schemavalidator" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/schemavalidator; fi
RUN if [ "$build_vulnerabilityreport" = "true" ]; then go build -o /app/out/plugins/ /app/plugins/verifier/vulnerabilityreport; fi

FROM $BASEIMAGE
LABEL org.opencontainers.image.source https://github.com/deislabs/ratify

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

