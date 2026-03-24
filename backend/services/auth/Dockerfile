# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./services/auth/cmd/api

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/auth-service .
COPY --from=builder /app/services/auth/config ./config
EXPOSE 8081
CMD ["./auth-service"]
