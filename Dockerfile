FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /clash-sub-aggregator .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget gzip && \
    ARCH=$(case "$(uname -m)" in x86_64) echo "amd64";; aarch64) echo "arm64";; esac) && \
    wget -O /tmp/mihomo.gz "https://github.com/MetaCubeX/mihomo/releases/download/v1.19.0/mihomo-linux-${ARCH}-v1.19.0.gz" && \
    gunzip /tmp/mihomo.gz && \
    mv /tmp/mihomo /usr/local/bin/mihomo && \
    chmod +x /usr/local/bin/mihomo

WORKDIR /app
COPY --from=builder /clash-sub-aggregator .
COPY configs/app.yaml configs/app.yaml

RUN mkdir -p data

EXPOSE 8080 7890 7891 9090

CMD ["./clash-sub-aggregator"]
