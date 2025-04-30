# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache git
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/wallet-service .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./wallet-service"]