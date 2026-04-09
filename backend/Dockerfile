# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/tradekai ./cmd/server

FROM busybox:1.36.1-musl AS busybox

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/tradekai /tradekai
COPY --from=busybox /bin/busybox /busybox

USER nonroot:nonroot

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
	CMD ["/busybox", "wget", "-qO-", "http://127.0.0.1:8080/api/v1/health"]

ENTRYPOINT ["/tradekai"]
