FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o web-bff ./services/web-bff/cmd/api

FROM alpine:3.19
RUN addgroup -S app && adduser -S app -G app
WORKDIR /home/app
COPY --from=builder --chown=app:app /app/web-bff .
USER app
EXPOSE 8080
CMD ["./web-bff"]
