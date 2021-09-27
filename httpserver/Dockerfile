FROM golang:alpine3.14 as builder

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/ /app/cmd/hora 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/referrerstore/ociregistry
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/verifier/nv2verifier
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/verifier/sbom
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/plugins/ /app/plugins/verifier/cosign

FROM alpine:3.14

RUN set -ex \
    && apk add --no-cache ca-certificates apache2-utils

ARG HORA_FOLDER=$HOME/.hora/
COPY --from=builder /app/out/hora /app/
COPY --from=builder /app/out/plugins ${HORA_FOLDER}/plugins
COPY --from=builder /app/config/config.json ${HORA_FOLDER}
ENV HORA_CONFIG=${HORA_FOLDER}
EXPOSE 6001
CMD ["./app/hora", "serve", "--http", ":6001"]