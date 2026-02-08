# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server

# Runtime stage
FROM alpine:latest

# Tambahkan keamanan extra
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S pahlawan && adduser -S pahlawan -G pahlawan

WORKDIR /home/pahlawan/

# Copy binary
COPY --from=builder /app/server .

# Jalankan sebagai user biasa, bukan root
USER pahlawan

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

CMD ["./server"]
