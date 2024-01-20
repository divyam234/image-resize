FROM --platform=${BUILDPLATFORM:-linux/amd64} tonistiigi/xx AS xx
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:alpine as builder

COPY --from=xx / /

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache pkgconfig

RUN /xx-apk --no-cache gcc musl-dev vips-dev

COPY go.* ./

RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=1 /xx-go build -o main &&  /xx-verify main

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine

RUN apk add --no-cache vips-dev

WORKDIR /app

RUN adduser -D -u 1000 user

COPY --chown=user:user --from=builder /main /app/main

USER user

ENV MALLOC_ARENA_MAX 2

ENTRYPOINT ["/app/main" ]