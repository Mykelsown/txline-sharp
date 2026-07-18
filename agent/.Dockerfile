# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o txline-agent .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# ca-certificates needed for HTTPS calls to TxLINE and Anthropic APIs
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/txline-agent .

# Data directory for signals.jsonl and arena_results.json
RUN mkdir -p /data

EXPOSE 8081

# credentials.json and .env are mounted as volumes at runtime,
# never baked into the image.
ENTRYPOINT ["./txline-agent"]
