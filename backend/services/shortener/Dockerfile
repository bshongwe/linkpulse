# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o shortener-service ./services/shortener/cmd/api

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app
WORKDIR /home/app
COPY --from=builder --chown=app:app /app/shortener-service .
COPY --from=builder --chown=app:app /app/services/shortener/config ./config
USER app
EXPOSE 8082
CMD ["./shortener-service"]
