# syntax=docker/dockerfile:1

# Build stage: compile the Go application with the requested toolchain.
FROM golang:1.22-bullseye AS builder

ARG GO_TOOLCHAIN=1.24.1

WORKDIR /app

# Install and download the specific Go toolchain specified in go.mod.
RUN go install "golang.org/dl/go${GO_TOOLCHAIN}@latest" && \
    /root/go/bin/go${GO_TOOLCHAIN} download

# Cache module downloads.
COPY go.mod go.sum ./
RUN /root/go/bin/go${GO_TOOLCHAIN} mod download

# Copy the entire project and build a statically linked binary.
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux /root/go/bin/go${GO_TOOLCHAIN} build -o /app/bin/cottage-manager ./

# Runtime stage: minimal image containing only the binary.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/bin/cottage-manager /app/cottage-manager

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/app/cottage-manager"]
