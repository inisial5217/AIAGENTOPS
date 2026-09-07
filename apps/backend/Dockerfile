# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# install build dependencies
RUN apk add --no-cache git ca-certificates

# cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# build stripped binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/server ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/server /app/server
COPY --from=builder /build/migrations /app/migrations

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/server"]
