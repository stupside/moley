# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git wget ca-certificates
WORKDIR /app

# Accept LDFLAGS as build argument
ARG LDFLAGS

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux sh -c "go build $LDFLAGS -o moley ."

# Download cloudflared in same stage
RUN wget -O /tmp/cloudflared https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 && \
    chmod +x /tmp/cloudflared

# Final stage - minimal with CA certificates
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root

# Copy binaries
COPY --from=builder /app/moley /usr/local/bin/moley
COPY --from=builder /tmp/cloudflared /usr/local/bin/cloudflared

ENTRYPOINT ["moley"]
CMD ["--help"]
