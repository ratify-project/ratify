FROM golang:alpine3.14 as builder

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/ /app/cmd/ratify 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/referrerstore/ociregistry
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/verifier/nv2verifier
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/verifier/sbom
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/verifier/cosign

FROM alpine:3.14

RUN set -ex \
    && apk add --no-cache ca-certificates apache2-utils

ARG RATIFY_FOLDER=$HOME/.ratify/
COPY --from=builder /app/out/ratify /app/
COPY --from=builder /app/out/plugins ${RATIFY_FOLDER}/plugins
COPY --from=builder /app/config/config.json ${RATIFY_FOLDER}
ENV RATIFY_CONFIG=${RATIFY_FOLDER}
EXPOSE 6001
CMD ["./app/ratify", "serve", "--http", ":6001"]