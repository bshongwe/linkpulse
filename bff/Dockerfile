FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o bff ./cmd/api

# Final stage
FROM alpine:3.19

# Install runtime dependencies and create non-root user
RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -S app && adduser -S app -G app

WORKDIR /home/app

# Copy binary from builder
COPY --from=builder --chown=app:app /build/bff .

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -fsS http://localhost:8080/health || exit 1

CMD ["./bff"]
