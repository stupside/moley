# Manual Docker build. Expects a pre-built `moley` binary in the build context
# (run `go build -o moley .` first, or use the artifact produced by GoReleaser).
# The published image at ghcr.io/stupside/moley is built by ko via GoReleaser,
# not by this file — see .goreleaser.yml.

# Keep GO_VERSION in sync with the `go` directive in go.mod.
ARG GO_VERSION=1.26.1
ARG CLOUDFLARED_VERSION=2026.3.0

# Builder stage for cloudflared
FROM golang:${GO_VERSION}-alpine3.22 AS cloudflared
ARG CLOUDFLARED_VERSION

# Install build dependencies
RUN apk --no-cache add \
    git \
    make \
    gcc \
    musl-dev

# Clone cloudflared at a pinned release tag for reproducible builds
RUN git clone --depth 1 --branch "${CLOUDFLARED_VERSION}" \
    https://github.com/cloudflare/cloudflared.git /go/src/cloudflared

# Set the working directory for cloudflared build
WORKDIR /go/src/cloudflared

# Build cloudflared
RUN make cloudflared

# Final runtime stage
FROM alpine:3.22.1 AS runtime

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Copy cloudflared binary from builder stage
COPY --from=cloudflared /go/src/cloudflared/cloudflared /usr/local/bin/cloudflared

# Copy the pre-built moley binary from the build context
COPY moley /usr/local/bin/moley

# Make binaries executable (COPY already sets root:root ownership)
RUN chmod +x /usr/local/bin/cloudflared /usr/local/bin/moley

# Runs as root; $HOME defaults to /root, which matches the volume mount
# paths used by docker-compose.yml (~/.moley → /root/.moley).
WORKDIR /root

ENTRYPOINT ["moley"]
CMD ["--help"]
