# GoReleaser Dockerfile - uses pre-built binaries from build context

# Builder stage for cloudflared
FROM golang:1.24.5-alpine3.22 AS cloudflared

# Install build dependencies
RUN apk --no-cache add \
    git \
    make \
    gcc \
    musl-dev

# Clone cloudflared repository
RUN git clone --depth 1 https://github.com/cloudflare/cloudflared.git /go/src/cloudflared

# Set the working directory for cloudflared build
WORKDIR /go/src/cloudflared

# Build cloudflared from latest
RUN make cloudflared

# Final runtime stage
FROM alpine:latest AS runtime

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
