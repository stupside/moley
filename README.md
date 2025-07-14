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

Moley is the easiest way to share your local development services using Cloudflare Tunnels and your own custom domains. Forget about reverse proxies, manual DNS, or complex infrastructure—Moley automates everything for you, so you can focus on building, not configuring.

## Why use Moley?

- **No extra infrastructure:** No need for Nginx, load balancers, or public servers.
- **Automatic DNS:** Instantly get public URLs for your local apps, with DNS records managed for you.
- **Professional presentation:** Use your own domain names for demos, APIs, or webhooks.
- **One config, one command:** Centralized YAML config and a single command to go live.
- **Secure by default:** Built on Cloudflare Tunnels, with best practices for token and config security.

Unlike traditional approaches that require setting up reverse proxies like Nginx Proxy Manager, Moley automatically creates and manages DNS records via the Cloudflare API. This means no additional infrastructure setup—just configure your local services and let Moley handle the rest.

| Approach                | Infrastructure Required | DNS Management      | Setup Complexity    |
|-------------------------|------------------------|---------------------|---------------------|
| **Moley**               | None                   | Automatic via API   | Single command      |
| cloudflared + Nginx     | Nginx Proxy Manager    | Manual dashboard    | Multiple services   |
| Manual Cloudflare       | None                   | Manual dashboard    | Complex configuration |

Moley eliminates the need for reverse proxies while providing automated DNS management - the best of both worlds.

## Key Benefits

Moley offers automated setup, streamlined tunnel creation and configuration, and automatic subdomain management. It handles resource cleanup on tunnel termination and supports exposing multiple local applications at once. You can use your own domain names for a professional presentation, and there is no need for reverse proxies, load balancers, or additional services. All configuration is centralized and validated, and operational logging is structured for clarity.

---

## Getting Started

Get your local app online in minutes:

```sh
# 1. Install cloudflared and authenticate
cloudflared tunnel login

# 2. Install Moley
brew install --cask stupside/tap/moley

# 3. Set your Cloudflare API token
moley config --cloudflare.token="your-api-token"

# 4. Initialize and run
moley tunnel init
moley tunnel run
```

### Docker Installation

```sh
# 1. Up
docker compose up -d

# 2. Run cloudflared and moley
task docker:cloudflared -- tunnel login
task docker:moley -- config --cloudflare.token="your-api-token"
task docker:moley -- tunnel init
task docker:moley -- tunnel run
```

> You can also use `docker compose run` to manually run commands. See examples above.

For a full list of available commands and options, run `moley --help` or `moley <command> --help`.

---

## Configuration Reference

### Configuration Schema

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `zone` | string | The DNS zone (your domain) to use for the tunnel | Yes |
| `apps` | array | Array of application configurations | Yes |
| `apps[].target.port` | integer | Local service port (1-65535) | Yes |
| `apps[].target.hostname` | string | Local hostname (typically `localhost`) | Yes |
| `apps[].expose.subdomain` | string | Public subdomain (e.g., `api` becomes `api.yourdomain.com`) | Yes |

---

## Security Considerations

**Important**: Never commit API tokens to version control. Use the `moley config --cloudflare.token` command for sensitive data.

Refer to [SECURITY.md](SECURITY.md) for comprehensive security guidelines.

---

## Troubleshooting

### Common Issues

**Authentication Errors**
- Verify cloudflared authentication: `cloudflared tunnel login`
- Confirm API token permissions (Zone:Read, DNS:Edit)
- Set API token: `moley config --cloudflare.token="your-token"`

**Configuration Errors**
- Validate `moley.yml` file structure
- Verify zone name format and ownership
- Ensure port numbers are within valid range (1-65535)
- Confirm subdomain specifications

**Network Issues**
- Verify local service availability
- Check firewall configurations
- Ensure port accessibility

**File System Errors**
- Run `moley tunnel init` to create configuration file
- Verify working directory contains `moley.yml`

---

## Architecture

### Core Components

- **Tunnel Manager**: Handles tunnel lifecycle and deployment
- **Runner Service**: Manages tunnel execution and monitoring
- **Configuration Generator**: Creates Cloudflare tunnel configurations
- **API Client**: Cloudflare API integration wrapper
- **Configuration Manager**: Configuration loading and validation

### Operational Flow

1. Configuration loading and validation from `moley.yml`
2. Tunnel creation or reuse of existing tunnel
3. DNS record configuration for specified applications
4. Cloudflare tunnel configuration generation
5. Tunnel service execution and monitoring
6. Resource cleanup on termination

---

## Development

### Prerequisites

Install [Task](https://taskfile.dev) - a modern alternative to Make:

```bash
# macOS
brew install go-task/tap/go-task

# Linux/Windows
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d

# Or see https://taskfile.dev/installation/ for other options
```

### Build Commands

```bash
# Build the application
task go:build

# Install globally
task go:install

# Run the application
task go:run
```

### Testing

```bash
# Run test suite
task go:test

# Generate coverage report
task go:coverage
```

### Code Quality

```bash
# Format code
task go:fmt

# Static analysis
task go:vet
```

### Available Tasks

See all available tasks:

```bash
task --list
```

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with appropriate tests
4. Ensure all tests pass
5. Submit a pull request

---

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:

1. Review the troubleshooting section
2. Consult the security documentation
3. Open an issue on GitHub
