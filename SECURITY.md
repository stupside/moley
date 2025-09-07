# Security Guidelines

This document outlines security best practices and considerations when using Moley.

## Configuration Security

### File Permissions and Storage

Moley uses two types of configuration files with different security requirements:

1. **Global Configuration** (`~/.moley/config.yml`)
   - **HIGHLY CONFIDENTIAL** - Contains API tokens
   - Automatically created with restrictive permissions (0600)
   - Never commit to version control
   - Store securely and back up separately from code

2. **Local Configuration** (`moley.yml`)
   - **Safe to commit** - Contains only tunnel and ingress configuration
   - No sensitive data (tunnel IDs are not secrets)
   - Can be shared publicly

3. **State File** (`moley.lock`)
   - Contains resource state tracking information
   - Safe to commit - helps with reproducible deployments
   - Used for crash recovery and incremental updates

### API Token Management

```bash
# Set tokens securely - never hardcode them
moley config set --cloudflare.token="your-api-token"

# Tokens are stored in ~/.moley/config.yml with restricted file permissions
```

**Critical Security Requirements:**
- Use Cloudflare API tokens with **minimum required permissions**
- Create zone-specific tokens when possible (avoid global account tokens)
- Rotate tokens regularly (recommended: every 90 days)
- Never commit tokens to version control
- Use separate tokens for development vs production environments

### Required Cloudflare Permissions

Your API token needs these specific permissions:

- `Zone:Zone:Read` - To query zone information
- `Zone:DNS:Edit` - To create/delete DNS records for tunnels
- `Cloudflare Tunnel:Edit` - To create/manage tunnels

## Runtime Security

### Resource Management

Moley implements several security-focused features:

- **Idempotent Operations**: Resources are only created when needed, preventing unintended duplicates
- **State Validation**: All resource operations check current state before making changes
- **Graceful Cleanup**: Proper teardown of resources on shutdown (Ctrl+C)
- **Dry-Run Mode**: Test configurations without creating real resources (`--dry-run`)

### File System Security

- Tunnel configuration files are stored in `~/.moley/tunnels/` with appropriate permissions
- Cloudflared credentials are stored in `~/.cloudflared/` (managed by cloudflared itself)
- Log files do not contain sensitive information

## Network Security

### Tunnel Security Model

Cloudflare Tunnels provide enterprise-grade security:

- **No Inbound Firewall Rules**: No ports need to be opened on your local machine
- **Encrypted Transit**: All traffic encrypted via TLS between your server and Cloudflare
- **DDoS Protection**: Automatic DDoS mitigation through Cloudflare's network
- **Origin IP Protection**: Your server's IP address is never exposed

### Local Service Considerations

- Configure your local services to bind to `localhost` only when possible
- Use HTTPS for your local services when handling sensitive data
- Consider implementing authentication even for development services
- Monitor which ports you're exposing through Moley

## Operational Security

### Access Control

- **DNS Zone Control**: Ensure only authorized users can modify your DNS zone
- **Tunnel Management**: Limit who has access to create/modify tunnels in your Cloudflare account
- **Configuration Management**: Control access to the global configuration file

### Monitoring and Auditing

```bash
# Use structured logging to monitor operations
moley --log-level=debug tunnel run

# Monitor Cloudflare dashboard for tunnel activity
# Review DNS record changes in your zone
```

### Incident Response

If you suspect compromise:

1. **Immediate**: Revoke the compromised API token in Cloudflare dashboard
2. **Clean Up**: Stop all Moley instances and clean up tunnels manually if needed
3. **Rotate**: Generate new API tokens with minimum required permissions
4. **Audit**: Review tunnel and DNS record changes in Cloudflare logs
5. **Update**: Change any exposed credentials or services

## Development Security

### Safe Development Practices

```bash
# Always use dry-run mode when testing configurations
moley tunnel run --dry-run

# Use separate Cloudflare accounts/zones for development
# Consider using separate API tokens for different environments
```

### Environment Isolation

- Use separate DNS zones for development vs production
- Create environment-specific tunnel configurations
- Never test with production API tokens or zones

## Supply Chain Security

### Dependencies

Moley uses minimal external dependencies:

- **cloudflared**: Official Cloudflare binary (required)
- **Go standard library**: Minimal attack surface
- **Cloudflare Go SDK**: Official SDK for API access

### Verification

- Always download Moley from official releases
- Verify checksums when available
- Use official Cloudflare binaries for cloudflared

## Troubleshooting Security Issues

### Authentication Problems

1. Check API token permissions in Cloudflare dashboard
2. Verify token is correctly set: ensure `~/.moley/config.yml` exists and contains your token
3. Test token with minimal operations first
4. Check cloudflared authentication: `cloudflared tunnel login`

### Permission Errors

- Ensure `~/.moley/` directory has proper permissions (755)
- Check that config file has restrictive permissions (600)
- Verify your Cloudflare account has necessary zone/tunnel permissions

### Network Issues

- Test connectivity to Cloudflare APIs
- Verify firewall rules allow outbound HTTPS connections
- Check that local services are accessible from localhost 