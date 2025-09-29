<p align="center">
  <img src=".github/images/moley.png" alt="Moley Logo" width="200"/><br/>
</p>

<p align="center">
  <a href="https://github.com/stupside/moley/releases/latest">
    <img src="https://img.shields.io/github/v/release/stupside/moley?style=flat-square" alt="Latest Release">
  </a>
  <a href="https://pkg.go.dev/github.com/stupside/moley">
    <img src="https://img.shields.io/badge/Go-Reference-00ADD8?style=flat-square&logo=go" alt="Go Reference">
  </a>
  <a href="https://github.com/stupside/homebrew-tap/blob/main/Casks/moley.rb">
    <img src="https://img.shields.io/badge/Homebrew-Available-FBB040?style=flat-square&logo=homebrew" alt="Homebrew">
  </a>
  <a href="https://github.com/stupside/moley/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/stupside/moley?style=flat-square" alt="License">
  </a>
  <a href="https://github.com/stupside/moley/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/stupside/moley/ci.yml?style=flat-square" alt="Build Status">
  </a>
</p>

# Moley

**Expose your localhost applications to the internet using Cloudflare Tunnels.**

Moley is a powerful CLI tool that simplifies exposing your localhost applications to the internet using [Cloudflare Tunnels](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/). With Moley, you can share your local development servers on your own custom domain without complex port forwarding or expensive tunnel services.

## 💰 Why Cloudflare Tunnels?

**Free Professional Infrastructure** - While other tunnel services charge $5-20+/month or offer limited free tiers, Cloudflare Tunnels are **completely free with full features**. Combined with a personal domain (~€10/year), you get professional-grade capabilities that would cost significantly more with alternatives:

| Feature | Cloudflare Tunnels + Domain |
|---------|----------------------------|
| **Annual Cost** | ~€10/year (domain only) |
| **Custom Domains** | ✅ Your own domain |
| **DDoS Protection** | ✅ Robust protection included |
| **Reliability** | ✅ Enterprise-grade uptime |
| **Zero Trust Security** | ✅ Cloudflare Access |
| **Bandwidth** | ✅ Unlimited |
| **SSL/TLS** | ✅ Automatic certificates |

**Bottom line**: For less than €1/month, you get what other services charge €20+/month for. Cloudflare Tunnels combined with [Zero Trust features](https://developers.cloudflare.com/cloudflare-one/) provide enterprise-grade security and performance that would require multiple paid subscriptions elsewhere.

## ✨ Key Benefits

### 💰 **Cost-Effective**
No monthly subscriptions. Just your domain (~€10/year).

### 🔐 **Zero Trust Security**
Built-in access control, authentication, and user management through Cloudflare's platform.

### 🏢 **Enterprise Infrastructure**
DDoS protection, automatic SSL certificates, and reliable uptime.

### ⚡ **Simple Setup**
One command to expose localhost. No complex configuration or manual DNS management.

## 📦 Installation

### Homebrew (Recommended)
```bash
brew install --cask stupside/tap/moley
```

### Go Install
```bash
go install github.com/stupside/moley@latest
```

### Manual Download
Download the latest binary from the [releases page](https://github.com/stupside/moley/releases/latest).

## 🚀 Quick Start

1. **Authentication**
   ```bash
   # Authenticate cloudflared with your account
   cloudflared tunnel login

   # Option 1: Set API token in config file
   moley config set --cloudflare.token="your-api-token"

   # Option 2: Use environment variable (recommended for CI/CD)
   export MOLEY_CLOUDFLARE_TOKEN="your-api-token"
   ```

2. **Initialize your project**
   ```bash
   moley tunnel init
   ```

3. **Configure your apps**
   ```bash
   # Option 1: Edit the generated moley.yml file
   # Option 2: Use environment variables (great for containers/CI)
   export MOLEY_TUNNEL_INGRESS_ZONE="yourdomain.com"
   export MOLEY_TUNNEL_INGRESS_APPS_0_TARGET_PORT="8080"
   export MOLEY_TUNNEL_INGRESS_APPS_0_EXPOSE_SUBDOMAIN="api"
   ```

4. **Start tunneling**
   ```bash
   # Foreground mode
   moley tunnel run

   # Background mode
   moley tunnel run --detach

   # Or run with everything configured via environment variables
   MOLEY_CLOUDFLARE_TOKEN="token" MOLEY_TUNNEL_INGRESS_ZONE="yourdomain.com" moley tunnel run
   ```

Your app is now accessible at `https://api.yourdomain.com`! 🎉

> 💡 **Pro tip**: Environment variables take precedence over config files and are perfect for CI/CD, Docker containers, and keeping secrets secure.

## 📚 Documentation

For complete documentation including configuration options, troubleshooting, and advanced usage, visit our [documentation site](https://stupside.github.io/moley).

### Quick Links

- 📖 **[Installation Guide](https://stupside.github.io/moley/docs/installation/)** - Detailed installation instructions
- ⚡ **[Quick Start](https://stupside.github.io/moley/docs/quick-start/)** - Get up and running in minutes
- ⚙️ **[Configuration](https://stupside.github.io/moley/docs/configuration/)** - Advanced configuration options
- 🔧 **[Troubleshooting](https://stupside.github.io/moley/docs/troubleshooting/)** - Common issues and solutions

## 🛠️ Development

### Prerequisites

- Go 1.23+
- [Cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) installed

### Building from Source

```bash
git clone https://github.com/stupside/moley.git
cd moley
go mod download
go build -o moley .
```

### Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cloudflare](https://cloudflare.com) for providing the tunnel infrastructure
- [Cloudflared](https://github.com/cloudflare/cloudflared) team for the tunnel client
