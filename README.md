# Moley - Cloudflare Tunnel Manager

A simple CLI tool for exposing local services through Cloudflare Tunnel with your own domain names.

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

# Run with debug logging
MOLEY_DEBUG=true ./moley run
```

## Security

**Important**: Never commit API tokens to version control. Always use environment variables for sensitive configuration.

See [SECURITY.md](SECURITY.md) for detailed security best practices.

## Configuration Validation

Moley performs comprehensive validation of your configuration:

- **Domain Validation**: Ensures valid domain names
- **Port Validation**: Validates port numbers (1-65535)
- **Subdomain Validation**: Ensures valid subdomain names
- **API Token Validation**: Validates Cloudflare API token presence
- **App Configuration**: Validates all app configurations

## Logging

Moley uses structured logging with the following levels:

- **Info**: General operational information
- **Warn**: Warning messages for non-critical issues
- **Error**: Error messages for critical issues
- **Debug**: Detailed debugging information

### Log Format

```json
{
  "level": "info",
  "time": "2024-01-01T12:00:00Z",
  "message": "Tunnel deployment completed successfully",
  "tunnel": "moley-1704110400000000000"
}
```

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

### Running Tests

```bash
go test ./...
```

### Code Quality

The codebase follows Go best practices:

- **Error Handling**: Comprehensive error handling with context
- **Validation**: Input validation at all levels
- **Logging**: Structured logging throughout
- **Testing**: Comprehensive test coverage
- **Documentation**: Clear documentation and comments

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Ensure cloudflared is authenticated: `cloudflared tunnel login`
   - Verify API token has correct permissions
   - Check API token is set in environment

2. **Configuration Errors**
   - Validate your `moley.yml` file
   - Check domain name format
   - Ensure ports are valid (1-65535)

3. **Network Issues**
   - Verify local services are running
   - Check firewall settings
   - Ensure ports are accessible

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
MOLEY_DEBUG=true ./moley run
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review the security documentation
3. Open an issue on GitHub