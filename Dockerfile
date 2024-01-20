FROM --platform=${BUILDPLATFORM} tonistiigi/xx AS cc-helpers
FROM --platform=${BUILDPLATFORM} golang:alpine as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add clang lld libcap-utils

COPY --from=cc-helpers / /

RUN xx-apk add --no-cache gcc musl-dev vips-dev

ENV CGO_ENABLED=1

WORKDIR /var/task

COPY go.* ./

RUN go mod download

COPY main.go ./

RUN xx-go build -o main &&  xx-verify main

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine

RUN apk add --no-cache vips-dev

WORKDIR /app

RUN adduser -D -u 1000 user

COPY --chown=user:user --from=builder /var/task/main /app/main

USER user

ENV MALLOC_ARENA_MAX 2

ENTRYPOINT ["/app/main" ]