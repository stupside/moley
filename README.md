<p align="center">
  <img src=".github/images/moley.png" alt="Moley Logo" width="200"/><br/>
</p>

<div style="display:flex;justify-content:center;gap:8px">
  <a href="https://pkg.go.dev/github.com/stupside/moley">
    <img src="https://pkg.go.dev/badge/github.com/stupside/moley.svg" alt="Go Reference">
  </a>
  <a href="https://github.com/stupside/homebrew-tap/blob/main/Casks/moley.rb">
    <img src="https://img.shields.io/badge/homebrew-install-brightgreen.svg" alt="Homebrew">
  </a>
</div>

# Moley

**Expose your local apps to the world—securely, instantly, and with zero hassle.**

Moley is the easiest way to share your local development services using Cloudflare Tunnels and your own custom domains. Moley automates everything for you, so you can focus on building, not configuring.

## Why Choose Moley?

- **Zero Infra**: No reverse proxies, load balancers, or additional services needed
- **Instant Setup**: Get your local app online in minutes with a single command
- **Custom Domains**: Use your own domain names for a professional presentation
- **Secure by Default**: Uses Cloudflare's enterprise-grade security and DDoS protection
- **Automatic Recovery**: Robust state management with crash recovery and cleanup

---

### Key Benefits

**Automated Setup**: Moley handles tunnel creation, DNS configuration, and ingress routing automatically. **Streamlined Operations**: Single command deployment with automatic subdomain management. **Resource Management**: Proper cleanup on tunnel termination with crash recovery. **Multi-App Support**: Expose multiple local applications simultaneously. **Professional Presentation**: Use your own domain names for a polished appearance.

**Zero Additional Infrastructure**: No need for reverse proxies, load balancers, or additional services. **Centralized Configuration**: All settings managed through a single, validated configuration file. **Operational Clarity**: Structured logging provides clear operational insights and debugging information.

Unlike traditional approaches that require setting up reverse proxies like Nginx Proxy Manager, Moley automatically creates and manages DNS records via the Cloudflare API. This means no additional infrastructure setup—just configure your local services and let Moley handle the rest.

| Approach                | Infrastructure Required | DNS Management      | Setup Complexity |
|-------------------------|------------------------|---------------------|------------------|
| **Moley**               | None                   | Automatic via API   | Single command   |
| cloudflared + Nginx     | Nginx Proxy Manager    | Manual dashboard    | Multiple services|
| Manual Cloudflare       | None                   | Manual dashboard    | Complex config   |

Moley eliminates the need for reverse proxies while providing automated DNS management. Simplicity first.

---

## Quick Start

### 1. Prerequisites

- [Cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) installed and authenticated
- A Cloudflare account with a custom domain
- Go 1.23+ (for building from source)

### 2. Installation

```bash
# Install via Homebrew
brew install --cask stupside/tap/moley

# Or download from releases
# Or build from source (see Development section)
```

### 3. Authentication

```bash
# Authenticate with Cloudflare
cloudflared tunnel login

# Set your Cloudflare API token
moley config set --cloudflare.token="your-api-token"
```

### 4. Initialize Configuration

```bash
# Create a new tunnel configuration
moley tunnel init
```

### 5. Configure Your Apps

Edit the generated `moley.yml` file:

```yaml
tunnel:
  id: "a-unique-tunnel-id"

ingress:
  zone: "yourdomain.com"
  apps:
    - target:
        port: 8080
        hostname: "localhost"
      expose:
        subdomain: "api"
    - target:
        port: 3000
        hostname: "localhost"
      expose:
        subdomain: "web"
```

### 6. Start Your Tunnel

```bash
# Start the tunnel service
moley tunnel run

# Stop when done: press Ctrl-C (Moley will clean up gracefully)
```

Your apps are now accessible at:
- `https://api.yourdomain.com` → `localhost:8080`
- `https://web.yourdomain.com` → `localhost:3000`

---

## Architecture

Moley uses **Hexagonal Architecture** (Ports & Adapters) to keep business logic separate from external dependencies, making it maintainable and testable:

### Core Components

- **Domain Layer**: Business entities (`Tunnel`, `Ingress`, `AppConfig`) with validation
- **Application Layer**: Orchestration services that coordinate resource lifecycle
- **Ports**: Interfaces defining contracts (`TunnelService`, `DNSService`)
- **Adapters**: Cloudflare-specific implementations of the ports
- **Framework**: Resource management with state tracking and idempotent operations

### Key Features

- **Resource Management Framework**: Declarative resource lifecycle with `Up`/`Down`/`Status` operations
- **State Persistence**: `moley.lock` file tracks deployed resources for crash recovery
- **Idempotent Operations**: Resources are only created/destroyed when needed
- **Configuration System**: Global config (`~/.moley/config.yml`) + local config (`moley.yml`)
- **Dry-Run Support**: Test configurations without making real changes
- **Graceful Shutdown**: Proper cleanup on termination signals

### Architecture Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                           CLI Layer                             │
│  moley tunnel run | moley tunnel init | moley config set        │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────────────────────────────────────┐
│                    Application Layer                           │
│              Tunnel Service (orchestrates resources)           │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────────────────────────────────────┐
│                    Resource Framework                          │
│    Resource Manager -> Resource Handlers -> Resource Lock      │
│    ┌──────────────┐ ┌─────────────────┐ ┌──────────────────┐   │
│    │ Tunnel Create│ │ Tunnel Config   │ │ DNS Records      │   │
│    │ Handler      │ │ Handler         │ │ Handler          │   │
│    └──────────────┘ └─────────────────┘ └──────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────────────────────────────────────┐
│                      Ports (Interfaces)                        │
│              TunnelService      │      DNSService               │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────────────────────────────────────┐
│                 Cloudflare Adapters                            │
│     tunnel.go (cloudflared CLI) │ dns.go (Cloudflare API)      │
└─────────────────────────────────────────────────────────────────┘
```

### Resource Lifecycle

1. **Parse Configuration**: Load tunnel and ingress configuration from `moley.yml`
2. **Create Resources**: Generate resource specifications (tunnel, config, DNS records)
3. **Diff Management**: Compare desired state vs. current state in `moley.lock`
4. **Sequential Operations**: Execute resource operations in dependency order
5. **State Persistence**: Update lock file with actual resource states
6. **Graceful Cleanup**: Reverse-order teardown on shutdown

---

## Configuration Reference

### Configuration Files

Moley uses two configuration files:

1. **Global Configuration** (`~/.moley/config.yml`): Contains sensitive data like API tokens
2. **Local Configuration** (`moley.yml`): Project-specific tunnel and ingress settings

### Configuration Schema

| Field | Type | Description | Required | Validation |
|-------|------|-------------|----------|------------|
| `tunnel.id` | string | Unique identifier for the tunnel | Yes | Generated UUID |
| `ingress.zone` | string | The DNS zone (your domain) to use for the tunnel | Yes | Valid domain format |
| `ingress.apps` | array | Array of application configurations | Yes | At least one app |
| `apps[].target.port` | integer | Local service port | Yes | 1-65535 |
| `apps[].target.hostname` | string | Local hostname (typically `localhost`) | Yes | Valid hostname/IP |
| `apps[].expose.subdomain` | string | Public subdomain (e.g., `api` becomes `api.yourdomain.com`) | Yes | Alphanumeric + hyphens |

### Global Configuration (`~/.moley/config.yml`)

```yaml
cloudflare:
  token: "your-cloudflare-api-token"
```

### Local Configuration (`moley.yml`)

```yaml
tunnel:
  id: "1663c83d-8801-424f-b060-734882126071"

ingress:
  zone: "example.com"
  apps:
    - target:
        port: 8080
        hostname: "localhost"
      expose:
        subdomain: "api"
    - target:
        port: 3000
        hostname: "localhost"
      expose:
        subdomain: "web"
```

### State Management

Moley creates a `moley.lock` file that tracks deployed resources:

```json
{
  "resources": {
    "hash1": {
      "handler": "tunnel-create",
      "payload": "..."
    },
    "hash2": {
      "handler": "dns-record", 
      "payload": "..."
    }
  }
}
```

This lockfile enables:
- **Crash Recovery**: Resume operations after interruption
- **Incremental Updates**: Only deploy changes, not everything
- **Clean Teardown**: Remove only what was actually created

---

## Commands

### Global Information

```bash
# Show version and build information
moley --version
moley info

# Adjust logging level
moley --log-level=debug tunnel run
```

### Configuration Management

```bash
# Set Cloudflare API token (stored in ~/.moley/config.yml)
moley config set --cloudflare.token="your-token"
```

### Tunnel Management

```bash
# Initialize a new tunnel configuration (creates moley.yml)
moley tunnel init

# Start the tunnel service (uses moley.yml by default)
moley tunnel run

# Use a custom configuration file
moley tunnel run --config custom-config.yml

```

### Available Commands

| Command | Description |
|---------|-------------|
| `moley info` | Display detailed build and configuration information |
| `moley config set` | Set global configuration values (API tokens, etc.) |
| `moley tunnel init` | Create a new tunnel configuration file with defaults |
| `moley tunnel run` | Deploy and run the tunnel with current configuration |
| `moley tunnel run --config FILE` | Deploy and run with a custom configuration file |
| `moley tunnel run --dry-run` | Validate configuration without creating resources |

### Development

```bash
# Run with dry-run mode (no actual resources created)
moley tunnel run --dry-run

# Use custom config file for different environments
moley tunnel run --config production.yml

# Use Task for development workflows
task go:test     # Run tests
task go:lint     # Run linting
task go:build    # Build binary
```

---

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/stupside/moley/v2.git
cd moley

# Install dependencies
go mod download

# Build the binary
go build -o moley .

# Or use Task for development
task go:build
```

### Docker Development

```bash
# Run moley in Docker
task docker:moley -- tunnel run

# Run cloudflared commands in Docker
task docker:cloudflared -- tunnel list
```

---

## Security

- **API Token Management**: Never commit API tokens to version control
- **Configuration Security**: Local configuration files contain sensitive data
- **Network Security**: All traffic encrypted via Cloudflare tunnels
- **Access Control**: Regular token rotation recommended

See [SECURITY.md](SECURITY.md) for detailed security guidelines.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
