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
FROM alpine:3.22.1 AS runtime

# Use buildx automatic platform detection for multi-arch builds
ARG TARGETARCH
ARG TARGETOS

# Install runtime dependencies and setup in a single layer
RUN apk --no-cache add ca-certificates && \
    adduser -D -s /bin/sh moley

# Copy cloudflared binary from builder stage
COPY --from=cloudflared /go/src/cloudflared/cloudflared /usr/local/bin/cloudflared

# Copy the pre-built binary from GoReleaser build context
# GoReleaser provides the binary directly in the build context
COPY moley /usr/local/bin/moley

# Make binaries executable and set proper ownership in a single layer
RUN chmod +x /usr/local/bin/cloudflared /usr/local/bin/moley && \
    chown root:root /usr/local/bin/cloudflared /usr/local/bin/moley

# Set working directory
WORKDIR /usr/local/bin

# Switch to non-root user for security
USER moley

ENTRYPOINT ["moley"]
CMD ["--help"]
