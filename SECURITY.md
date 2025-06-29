# Security Best Practices

## API Token Security

### Never commit API tokens to version control
- Always use environment variables for sensitive configuration
- Set your Cloudflare API token via the `MOLEY_CLOUDFLARE_API_TOKEN` environment variable
- Never hardcode tokens in configuration files

### Example usage:
```bash
export MOLEY_CLOUDFLARE_API_TOKEN="your-api-token-here"
moley run
```

### Or inline:
```bash
MOLEY_CLOUDFLARE_API_TOKEN="your-api-token-here" moley run
```

## Configuration Security

### Use environment variables for sensitive data
The application supports the following environment variables:
- `MOLEY_CLOUDFLARE_API_TOKEN` - Your Cloudflare API token

### Configuration file security
- Keep your `moley.yml` file secure and don't share it publicly
- Use `.gitignore` to prevent accidentally committing sensitive configuration
- Consider using separate configuration files for different environments

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
4. **Backup**: Keep secure backups of your configuration
5. **Documentation**: Document your tunnel setup and access procedures

## Troubleshooting

If you encounter authentication issues:
1. Verify your API token is valid and has the correct permissions
2. Check that the token is properly set in the environment
3. Ensure your Cloudflare account has the necessary permissions
4. Verify that cloudflared is properly authenticated with `cloudflared tunnel login` 