# CloudBoot BootOS - Lightweight Linux for bare-metal provisioning
# Based on Alpine Linux for minimal footprint

FROM alpine:3.19 AS builder

# Install build dependencies
RUN apk add --no-cache \
    go \
    gcc \
    musl-dev \
    git

# Build cb-agent
WORKDIR /build
COPY cb-agent /build/cb-agent
WORKDIR /build/cb-agent

RUN go build -ldflags="-s -w" -o /usr/local/bin/cb-agent ./main.go

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    bash \
    curl \
    dmidecode \
    lsblk \
    util-linux \
    pciutils \
    usbutils \
    iproute2 \
    ethtool \
    hdparm \
    smartmontools \
    kexec-tools

# Copy agent binary
COPY --from=builder /usr/local/bin/cb-agent /usr/local/bin/cb-agent

# Copy init script
COPY init.sh /init.sh
RUN chmod +x /init.sh

# Set environment variables
ENV CB_SERVER_URL=http://10.0.0.1:8080
ENV CB_POLL_INTERVAL=5s

# Entry point
ENTRYPOINT ["/init.sh"]
