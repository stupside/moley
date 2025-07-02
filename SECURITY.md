# Security Best Practices

## API Token Security

### Never commit API tokens to version control
- Always use the `moley config --cloudflare.token` command for sensitive configuration
- Never hardcode tokens in configuration files

### Example usage:
```bash
cloudflared tunnel login
moley config --cloudflare.token="your-api-token-here"
```

## Configuration Security

### Configuration file security
- The `moley.yml` file in this repository is **not confidential** and can be safely shared or committed
- The configuration file in your home directory (`~/.moley/config.yml`) **is confidential** and should be kept secure
- Use `.gitignore` to prevent accidentally committing sensitive configuration from your home directory
- Consider using separate configuration files for different environments if needed

## Network Security

### Tunnel security
- Cloudflare tunnels provide end-to-end encryption
- All traffic is encrypted in transit
- Use HTTPS for your local services when possible

### Access control
- Regularly rotate your Cloudflare API tokens
- Use the principle of least privilege for API token permissions
- Monitor tunnel usage and access logs

## Best Practices

1. **Regular updates**: Keep the application and dependencies updated
2. **Token rotation**: Regularly rotate your Cloudflare API tokens
3. **Monitoring**: Monitor tunnel usage and access patterns
4. **Backup**: Keep secure backups of your home directory configuration (`~/.moley/config.yml`)
5. **Documentation**: Document your tunnel setup and access procedures

## Troubleshooting

If you encounter authentication issues:
1. Verify your API token is valid and has the correct permissions
2. Check that the token is properly set using the `moley config` command
3. Ensure your Cloudflare account has the necessary permissions
4. Verify that cloudflared is properly authenticated with `cloudflared tunnel login` 