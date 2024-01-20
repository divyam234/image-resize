FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:alpine as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache pkgconfig git gcc musl-dev vips-dev

WORKDIR /var/task

COPY go.* ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o main

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine

RUN apk add --no-cache vips-dev

WORKDIR /app

RUN adduser -D -u 1000 user

COPY --chown=user:user --from=builder /var/task/main /app/main

USER user

ENV MALLOC_ARENA_MAX 2

ENTRYPOINT ["/app/main" ]