<p align="center">
  <img src=".github/images/moley.png" alt="Moley Logo" width="200"/>
</p>

# Moley - Cloudflare Tunnel Manager

A simple CLI tool for exposing local services through Cloudflare Tunnel with your own domain names.

## Why Moley?

I wanted to expose my localhost services with custom domain names without paying for services like ngrok. Cloudflare Tunnel is free and powerful, but setting it up manually can be complex. Moley simplifies this process into a few simple commands.

## What it solves

Moley is a wrapper around Cloudflare Tunnel that:
- **Creates tunnels automatically** - No manual setup required
- **Manages DNS records** - Automatically creates subdomains for your services
- **Handles cleanup** - Removes all resources when you're done
- **Works with multiple services** - Expose multiple local apps at once
- **Uses your own domain** - No more random URLs or subdomains

## Features

- 🚀 **Easy Setup**: Simple YAML configuration
- 🔒 **Secure**: End-to-end encrypted tunnels
- 🎯 **Domain Control**: Use your own domain names
- 🧹 **Auto Cleanup**: Automatic resource cleanup on exit
- 📝 **Structured Logging**: Comprehensive logging with structured fields
- ✅ **Validation**: Robust configuration validation
- 🔧 **Flexible**: Support for multiple applications

## Installation

### Prerequisites

1. **Cloudflare Account**: You need a Cloudflare account with a domain
2. **Cloudflare API Token**: Create an API token with Zone:Read and DNS:Edit permissions
3. **cloudflared**: Install and authenticate cloudflared

### Install cloudflared

```bash
# macOS
brew install cloudflare/cloudflare/cloudflared

# Linux
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb

# Windows
# Download from https://github.com/cloudflare/cloudflared/releases
```

### Authenticate cloudflared

```bash
cloudflared tunnel login
```

### Build Moley

```bash
git clone <repository-url>
cd mole
make build
```

## Configuration

Create a `moley.yml` configuration file:

```yaml
# Cloudflare configuration
cloudflare:
  api_token: ""  # Set via MOLEY_CLOUDFLARE_API_TOKEN environment variable

# Zone configuration
zone: yourdomain.com

# Applications to expose
apps:
  - port: 3000
    subdomain: api
  - port: 8080
    subdomain: web
```

### Environment Variables

- `MOLEY_CLOUDFLARE_API_TOKEN`: Your Cloudflare API token (recommended)

## Usage

### Basic Usage

```bash
# Set your API token
export MOLEY_CLOUDFLARE_API_TOKEN="your-api-token-here"

# Run the tunnel
./moley run
```

### Advanced Usage

```bash
# Run with inline environment variable
MOLEY_CLOUDFLARE_API_TOKEN="your-token" ./moley run
```

## Command-line Flags

Moley supports command-line flags for customizing its behavior and overriding configuration values directly from the CLI.

To discover all available flags and options for any command, use the built-in help system:

```bash
./moley --help
./moley <command> --help
```

This will display detailed information about available flags, their usage, and examples for each command.

## Security

**Important**: Never commit API tokens to version control. Always use environment variables for sensitive configuration.

See [SECURITY.md](SECURITY.md) for detailed security best practices.

## Configuration Validation

Moley performs comprehensive validation of your configuration:

- **Zone Validation**: Ensures zone is specified
- **Port Validation**: Validates port numbers (1-65535)
- **Subdomain Validation**: Ensures subdomain names are provided
- **API Token Validation**: Validates Cloudflare API token presence
- **App Configuration**: Validates all app configurations

## Logging

Moley uses structured logging with the following levels:

- **Info**: General operational information
- **Warn**: Warning messages for non-critical issues
- **Error**: Error messages for critical issues

## Architecture

### Components

- **Manager**: Handles tunnel deployment and cleanup
- **Runner**: Manages tunnel execution and lifecycle
- **Generator**: Creates Cloudflare tunnel configurations
- **Client**: Cloudflare API client wrapper

### Flow

1. **Configuration Loading**: Load and validate configuration
2. **Tunnel Creation**: Create or reuse existing tunnel
3. **DNS Setup**: Configure DNS records for applications
4. **Configuration Generation**: Generate cloudflared configuration
5. **Tunnel Execution**: Start and manage tunnel process
6. **Cleanup**: Clean up resources on exit

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Ensure cloudflared is authenticated: `cloudflared tunnel login`
   - Verify API token has correct permissions
   - Check API token is set in environment

2. **Configuration Errors**
   - Validate your `moley.yml` file
   - Check zone name format
   - Ensure ports are valid (1-65535)

3. **Network Issues**
   - Verify local services are running
   - Check firewall settings
   - Ensure ports are accessible

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review the security documentation
3. Open an issue on GitHub