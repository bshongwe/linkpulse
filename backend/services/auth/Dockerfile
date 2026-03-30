# ----------------------
# Build stage
# ----------------------
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./services/auth/cmd/api

# ----------------------
# Runtime stage
# ----------------------
FROM alpine:3.19
RUN addgroup -S app \
    && adduser -S app -G app
WORKDIR /home/app
COPY --from=builder --chown=app:app /app/auth-service .
COPY --from=builder --chown=app:app /app/services/auth/config ./config
USER app
EXPOSE 8081
CMD ["./auth-service"]
