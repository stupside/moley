# GoReleaser Dockerfile - uses pre-built binaries from build context
FROM alpine:latest

# Use buildx automatic platform detection for multi-arch builds
ARG TARGETARCH
ARG TARGETOS

# Install ca-certificates and wget for cloudflared
RUN apk --no-cache add ca-certificates wget

# Download and install cloudflared for the target architecture
RUN wget -O /usr/local/bin/cloudflared \
        "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-${TARGETOS}-${TARGETARCH}" && \
    chmod +x /usr/local/bin/cloudflared

# Set working directory
WORKDIR /root

# Copy the pre-built binary from GoReleaser build context
# GoReleaser will automatically provide the correct binary for each platform
COPY moley /usr/local/bin/moley
RUN chmod +x /usr/local/bin/moley

ENTRYPOINT ["moley"]
CMD ["--help"]
