FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/pagbest154-cmd/system-monitor/internal/version.Version=1.0.0" -o /out/system-monitor ./cmd/system-monitor

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
ENV SYSTEM_MONITOR_ROOT=/app \
    SYSTEM_MONITOR_CONFIG_DIR=/app/config \
    SYSTEM_MONITOR_DATA_DIR=/app/data
COPY --from=builder /out/system-monitor /app/system-monitor
COPY web/ /app/web/
COPY config/ /app/config/
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["/app/system-monitor", "--mode", "hub", "--host", "0.0.0.0", "--port", "8080", "--proxy-headers"]
