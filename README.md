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

# Set your Cloudflare API token (creates ~/.moley/config.yml)
moley config set --cloudflare.token="your-api-token"

# Or use environment variable instead of config file
export MOLEY_CLOUDFLARE_TOKEN="your-api-token"
```

This creates your global configuration file at `~/.moley/config.yml` containing sensitive data like API tokens. Alternatively, you can use environment variables which take precedence over file configuration.

### 4. Initialize Configuration

```bash
# Create a new tunnel configuration (auto-generates moley.yml with defaults)
moley tunnel init
```

This automatically creates a `moley.yml` file with a unique tunnel ID and example configuration.

### 5. Configure Your Apps

Edit the auto-generated `moley.yml` file to match your local services:

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
# Start the tunnel service in the foreground
moley tunnel run

# Or run in detached mode (background)
moley tunnel run --detach

# Stop the tunnel (works for both foreground and detached)
moley tunnel stop

# Test configuration without creating resources
moley tunnel run --dry-run

# Use a custom configuration file
moley tunnel run --config custom-config.yml
```

**Detached Mode**: When using `--detach`, the tunnel runs in the background and you get your terminal back immediately. The tunnel continues running even if you close your terminal. Use `moley tunnel stop` to stop it later.

**Smart Stop**: The stop command intelligently cleans up all resources (tunnels, DNS records, processes). It first uses the `moley.lock` file to remove tracked resources, then uses your `moley.yml` configuration to detect and remove any orphaned resources, even if the lock file is missing or corrupted.

Your apps are now accessible at:
- `https://api.yourdomain.com` → `localhost:8080`
- `https://web.yourdomain.com` → `localhost:3000`

### 7. Additional Commands

```bash
# View version and build information
moley --version
moley info

# Debug with verbose logging
moley --log-level=debug tunnel run

# Use environment variables for configuration
MOLEY_CLOUDFLARE_TOKEN="token" MOLEY_TUNNEL_INGRESS_ZONE="mydomain.com" moley tunnel run
```

---

## Configuration Reference

### Configuration Files

Moley uses two configuration files with environment variable override support:

1. **Local Configuration** (`moley.yml`): Project-specific tunnel and ingress settings (auto-generated by `moley tunnel init`)
   - Environment variables with `MOLEY_TUNNEL_` prefix override file values
2. **Global Configuration** (`~/.moley/config.yml`): Contains sensitive data like API tokens
   - Environment variables with `MOLEY_` prefix override file values

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

**Environment Variable Override:**
You can override any global configuration value using environment variables with the `MOLEY_` prefix:

```bash
# Override Cloudflare token
export MOLEY_CLOUDFLARE_TOKEN="your-api-token"
```

Environment variables take precedence over file configuration.

### Local Configuration (`moley.yml`)

Auto-generated by `moley tunnel init` with a unique tunnel ID:

```yaml
tunnel:
  id: "1663c83d-8801-424f-b060-734882126071"  # Auto-generated UUID

ingress:
  zone: "example.com"  # Edit to match your domain
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

**Environment Variable Override:**
You can override any tunnel configuration value using environment variables with the `MOLEY_TUNNEL_` prefix:

```bash
# Override tunnel zone
export MOLEY_TUNNEL_INGRESS_ZONE="mydomain.com"

# Override tunnel ID (useful for CI/CD)
export MOLEY_TUNNEL_TUNNEL_ID="custom-tunnel-id"
```

Environment variables take precedence over file configuration, making them ideal for CI/CD pipelines or containerized deployments.

### State Management

Moley maintains state through a `moley.lock` file to enable intelligent resource management:

**Why a lockfile?**
- **Diff Detection**: Compare current state vs desired configuration to only update what changed
- **Crash Recovery**: Resume operations after interruption without recreating existing resources
- **Smart Cleanup**: Know exactly what resources to remove during `moley tunnel stop`
- **Idempotent Operations**: Avoid duplicate resources by tracking what's already deployed
- **Incremental Updates**: When you modify `moley.yml`, only deploy the differences

**Lockfile Structure:**

```json
{
  "entries": [
    {
      "handler_name": "tunnel-create",
      "data": {
        "config": { "tunnel": { "id": "..." } },
        "state": { "tunnel": { "id": "..." } }
      }
    },
    {
      "handler_name": "tunnel-run",
      "data": {
        "config": { "tunnel": { "id": "..." } },
        "state": { "pid": 31124, "tunnel": { "id": "..." } }
      }
    },
    {
      "handler_name": "dns-record",
      "data": {
        "config": { "subdomain": "api", "zone": "example.com" },
        "state": { "record_id": "..." }
      }
    }
  ]
}
```

**Key Benefits:**
- **Efficient**: Only creates/updates/removes resources that actually changed
- **Reliable**: Survives crashes and interruptions without losing track of resources
- **Safe**: Prevents accidental duplicate resources or orphaned infrastructure
- **Transparent**: You can see exactly what Moley has deployed at any time

---

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/stupside/moley.git
cd moley

# Install dependencies
go mod download

# Build the binary
go build -o moley .

# Or use Task for development workflows
task go:build    # Build binary
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
