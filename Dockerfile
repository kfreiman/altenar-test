# syntax=docker/dockerfile:1
FROM golang:1.26.1-alpine3.22 AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.io,direct

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o casino .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder --chown=appuser:appgroup /app/casino .
USER appuser
EXPOSE 8080
ENTRYPOINT ["./casino"]
CMD ["api"]