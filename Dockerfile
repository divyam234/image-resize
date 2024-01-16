FROM golang:1.21-alpine as builder

RUN apk add pkgconfig git gcc musl-dev vips-dev

WORKDIR /var/task

COPY go.* ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main

FROM alpine

RUN apk add --no-cache vips-dev

WORKDIR /app

RUN adduser -D -u 1000 user

COPY --chown=user:user --from=builder /var/task/main /app/main

USER user

ENTRYPOINT ["MALLOC_ARENA_MAX=2","/app/main" ]