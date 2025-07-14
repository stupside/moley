# GoReleaser Dockerfile - uses pre-built binaries from build context
FROM alpine:latest

# Install ca-certificates and wget for cloudflared
RUN apk --no-cache add ca-certificates wget

# Use buildx automatic platform detection
ARG TARGETARCH

# Download and install cloudflared
RUN wget -O /usr/local/bin/cloudflared \
        "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${TARGETARCH}" && \
    chmod +x /usr/local/bin/cloudflared

WORKDIR /root

# Copy the pre-built binary from GoReleaser build context
COPY moley /usr/local/bin/moley

ENTRYPOINT ["moley"]
CMD ["--help"]
