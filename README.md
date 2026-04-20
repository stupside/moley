<p align="center">
  <img src=".github/images/moley.svg" alt="Moley" width="200"/>
</p>

<p align="center">
  <a href="https://github.com/stupside/moley/releases/latest"><img src="https://img.shields.io/github/v/release/stupside/moley?style=flat-square" alt="Release"></a>
  <a href="https://pkg.go.dev/github.com/stupside/moley"><img src="https://img.shields.io/badge/Go-Reference-00ADD8?style=flat-square&logo=go" alt="Go Reference"></a>
  <a href="https://github.com/stupside/homebrew-tap/blob/main/Casks/moley.rb"><img src="https://img.shields.io/badge/Homebrew-Available-FBB040?style=flat-square&logo=homebrew" alt="Homebrew"></a>
  <a href="https://github.com/stupside/moley/blob/main/LICENSE"><img src="https://img.shields.io/github/license/stupside/moley?style=flat-square" alt="License"></a>
  <a href="https://github.com/stupside/moley/actions"><img src="https://img.shields.io/github/actions/workflow/status/stupside/moley/ci.yml?style=flat-square" alt="CI"></a>
</p>

# Moley

Share a local app on your own domain in one command.

Moley unlocks the good parts of Cloudflare ([Tunnels](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/), DNS, [Access](https://developers.cloudflare.com/cloudflare-one/policies/access/)) straight from a single config file. Point it at `localhost:3000`, declare who can see it, get a live URL. You never have to touch the Cloudflare dashboard.

## What you can do

One `moley.yml`, one command, and Cloudflare does the rest. A few real setups:

### Share a dev app with a teammate

You're building something on `localhost:3000` and want to show it to one
person, signed in with their GitHub account, on your own domain.

```yaml
access:
  policies:
    - name: just-us
      decision: allow
      include:
        - email: {email: "teammate@example.com"}

ingress:
  zone: "example.com"
  mode: subdomain
  apps:
    - target: {hostname: localhost, port: 3000, protocol: http}
      expose: {subdomain: demo}
      access:
        providers: [github, onetimepin]
        session_duration: "168h"
      policies: [just-us]
```

`moley tunnel run` → `demo.example.com`, behind a GitHub login, locked to
the email you listed.

### Let the whole company in

Swap the policy. Anyone with a company email, signed in through your
corporate IdP, gets through.

```yaml
access:
  policies:
    - name: company-only
      decision: allow
      include:
        - email_domain: {domain: "mycompany.com"}

ingress:
  zone: "example.com"
  mode: subdomain
  apps:
    - target: {hostname: localhost, port: 3000, protocol: http}
      expose: {subdomain: staging}
      access:
        providers: [google-workspace]
      policies: [company-only]
```

### Gate an API for a single service

Expose a local API so only one caller can reach it, using a Cloudflare
service token. No browser flow, no IdP.

```yaml
access:
  policies:
    - name: one-client
      decision: non_identity
      include:
        - service_token: {token_id: "your-service-token-uuid"}

ingress:
  zone: "example.com"
  mode: subdomain
  apps:
    - target: {hostname: localhost, port: 8080, protocol: http}
      expose: {subdomain: api}
      policies: [one-client]
```

The caller sends `CF-Access-Client-Id` and `CF-Access-Client-Secret`
headers. Everyone else gets blocked at Cloudflare's edge, before a single
packet hits your laptop.

### Mix and match

Nothing stops you from combining these in one file. Dashboard behind a
GitHub login, webhook behind a service token, public status page behind
nothing at all, all running through the same tunnel, all cleaned up when
you run `moley tunnel stop`.

## Install

Homebrew:

```bash
brew install --cask stupside/tap/moley
```

Go:

```bash
go install github.com/stupside/moley/v2@latest
```

Or grab a binary from the [releases page](https://github.com/stupside/moley/releases/latest).

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

MIT. See [LICENSE](LICENSE).
