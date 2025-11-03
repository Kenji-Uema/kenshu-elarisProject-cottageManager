# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOTOOLCHAIN=go1.24.1 go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOTOOLCHAIN=go1.24.1 go build -o /bin/cottageManager .

FROM gcr.io/distroless/base-debian12 AS runtime

WORKDIR /app

COPY --from=builder /bin/cottageManager /app/cottageManager

EXPOSE 8080

ENV GIN_PORT=8080

USER nonroot:nonroot

ENTRYPOINT ["/app/cottageManager"]

