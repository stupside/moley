<p align="center">
  <img src=".github/images/moley.png" alt="Moley Logo" width="200"/>
</p>

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
# 1. Install cloudflared (choose your OS)
# macOS
brew install cloudflare/cloudflare/cloudflared
# Linux
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb
# Windows: Download from https://github.com/cloudflare/cloudflared/releases

# 2. Authenticate cloudflared with your Cloudflare account
cloudflared tunnel login

# 3. Clone and build Moley
git clone <repository-url>
cd mole
make build

# 4. Set your Cloudflare API token
./moley config --cloudflare.token="your-api-token"

# 5. Initialize Moley configuration
./moley tunnel init

# 6. Edit the generated moley.yml file to match your requirements
# (open moley.yml in your editor)

# 7. Start the tunnel
./moley tunnel run
```

For a full list of available commands and options, run `./moley --help` or `./moley <command> --help`.

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

### Build Commands

```bash
# Development build
make build

# Global installation
make install

# Clean build artifacts
make clean
```

### Testing

```bash
# Run test suite
make test

# Generate coverage report
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Static analysis
make vet
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