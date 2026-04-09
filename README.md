<p align="center">
  <img src=".github/images/moley.png" alt="Moley" width="200"/>
</p>

<p align="center">
  <a href="https://github.com/stupside/moley/releases/latest"><img src="https://img.shields.io/github/v/release/stupside/moley?style=flat-square" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/stupside/moley"><img src="https://img.shields.io/badge/Go-Reference-00ADD8?style=flat-square&logo=go" alt="Go Reference"></a>
  <a href="https://github.com/stupside/homebrew-tap/blob/main/Casks/moley.rb"><img src="https://img.shields.io/badge/Homebrew-Available-FBB040?style=flat-square&logo=homebrew" alt="Homebrew"></a>
  <a href="https://github.com/stupside/moley/blob/main/LICENSE"><img src="https://img.shields.io/github/license/stupside/moley?style=flat-square" alt="License"></a>
  <a href="https://github.com/stupside/moley/actions"><img src="https://img.shields.io/github/actions/workflow/status/stupside/moley/ci.yml?style=flat-square" alt="CI"></a>
</p>

# Moley

Moley automates [Cloudflare Tunnels](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) so you can expose local services on your own domain without manual setup.

It handles tunnel creation, DNS records, ingress config, and cleanup.

## Install

### Homebrew

```bash
brew install --cask stupside/tap/moley
```

### Go

```bash
go install github.com/stupside/moley/v2@latest
```

### Binary

Download from the [releases page](https://github.com/stupside/moley/releases/latest).

## Quick start

```bash
cloudflared tunnel login

moley config set --cloudflare.token="your-api-token"

moley tunnel init

moley tunnel run
```

## Documentation

Full documentation is available at [moley.dev](https://moley.dev).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT - see [LICENSE](LICENSE).
