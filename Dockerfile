# Builder stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ssl certificates (needed for SDK network calls)
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
# -ldflags="-w -s" strips debug information to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o databricks-connector cmd/databricks-connector/main.go

# Final stage
FROM alpine:3.24

# Upgrade packages to fix vulnerabilities (e.g. busybox) and install ca-certificates
RUN apk upgrade --no-cache && apk add --no-cache ca-certificates

# Copy the binary
COPY --from=builder /app/databricks-connector /databricks-connector

# Run as non-root user (ID 65532 is commonly used for nonroot)
USER 65532:65532

ENTRYPOINT ["/databricks-connector"]
