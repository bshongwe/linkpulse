# ----------------------
# Build stage
# ----------------------
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o analytics-service ./services/analytics/cmd/api

# ----------------------
# Runtime stage
# ----------------------
FROM alpine:3.19
RUN addgroup -S app \
    && adduser -S app -G app
WORKDIR /home/app
COPY --from=builder --chown=app:app /app/analytics-service .
USER app
EXPOSE 8082
CMD ["./analytics-service"]
