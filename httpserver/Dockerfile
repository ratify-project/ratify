ARG BUILDPLATFORM="linux/amd64"
ARG BUILDERIMAGE="golang:1.19"
ARG BASEIMAGE="gcr.io/distroless/static:nonroot"

FROM --platform=$BUILDPLATFORM $BUILDERIMAGE as builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG LDFLAGS
ARG GOPROXY

ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT} \
    GOPROXY=${GOPROXY}

WORKDIR /app

COPY . .

RUN go build -o /app/out/ /app/cmd/ratify 

FROM $BASEIMAGE
LABEL org.opencontainers.image.source https://github.com/deislabs/ratify

ARG RATIFY_FOLDER=$HOME/.ratify/

WORKDIR /

COPY --from=builder /app/out/ratify /app/
COPY --from=builder /app/config/config.json ${RATIFY_FOLDER}

ENV RATIFY_CONFIG=${RATIFY_FOLDER}

EXPOSE 6001
EXPOSE 8888

USER 65532:65532

ENTRYPOINT ["/app/ratify", "serve", "--http", ":6001"]

